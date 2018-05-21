package shh

/*

Simple External Poller Integration: If you can open a socket, you can write a poller.

Format: <RFC3339 date stamp> <what> <value>\n

The exact interpretation of these depends on the Outputter in use.

Example

In terminal A:
  SHH_POLLERS=listen ./shh

In a different terminal:
  (while true; do echo $(date "+%Y-%m-%dT%H:%M:%SZ") memfree $(grep MemFree /proc/meminfo | awk '{print $2}').0; sleep 5; done) | nc -U \#shh

*/

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/heroku/slog"
)

var (
	MetricNameRegexp = regexp.MustCompile("^[a-zA-Z0-9]([a-zA-Z0-9.-]+)?$")
	UnitRegexp       = regexp.MustCompile("^([a-zA-Z$%#]+)(,([a-zA-Z$%#]+))?$") // <unit 1>,<abbr 3>
)

type Listen struct {
	measurements chan<- Measurement
	listener     net.Listener
	Timeout      time.Duration
	metricCount,
	connectionCount,
	parseErrorCount uint64
	meta      bool
	closeDown chan struct{}
}

func NewListenPoller(measurements chan<- Measurement, config Config) Listen {
	ctx := slog.Context{"poller": "listen", "fn": "NewListenPoller"}
	tmp := strings.Split(config.Listen, ",")

	formatErr := fmt.Errorf("SHH_LISTEN is not in the correct format")
	fix := "The correct format is: <tcp|tcp4|tcp6|unix|unixpacket>,<address>"

	if len(tmp) != 2 {
		FatalError(ctx, formatErr, fix)
	}

	listenNet := tmp[0]
	listenLaddr := tmp[1]

	switch listenNet {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
		break
	default:
		FatalError(ctx, formatErr, fix)
	}

	// If this is a path, remove it
	if listenNet == "unix" && Exists(listenLaddr) {
		err := os.Remove(listenLaddr)
		if err != nil {
			FatalError(ctx, err, "unable to remove old socket path")
		}
	}

	listener, err := net.Listen(listenNet, listenLaddr)

	if err != nil {
		FatalError(ctx, err, "unable to listen on "+listenNet+listenLaddr)
	}

	poller := Listen{
		measurements: measurements,
		listener:     listener,
		Timeout:      config.ListenTimeout,
		closeDown:    make(chan struct{}, 1),
		meta:         config.Meta,
	}

	go poller.Accept()

	return poller
}

func (poller *Listen) Accept() {
	ctx := slog.Context{"poller": poller.Name(), "fn": "Accept"}
	for {
		conn, err := poller.listener.Accept()
		if err != nil {
			select {
			case <-poller.closeDown:
				LogError(ctx, err, "shutting down")
				return
			default:
				LogError(ctx, err, "accepting connection")
				continue
			}
		}

		go poller.HandleListenConnection(conn)
	}
}

func (poller Listen) Name() string {
	return "listen"
}

func (poller Listen) Exit() {
	poller.closeDown <- struct{}{}
	poller.listener.Close()
}

func (poller Listen) Poll(tick time.Time) {
	if poller.meta {
		poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{"_meta_", "metric", "count"}, poller.metricCount, Empty}
		poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{"_meta_", "connection", "count"}, poller.connectionCount, Empty}
		poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{"_meta_", "parse", "error", "count"}, poller.parseErrorCount, Empty}
	}
}

func (poller *Listen) HandleListenConnection(conn net.Conn) {
	defer conn.Close()

	ctx := slog.Context{"poller": poller.Name(), "fn": "handleListenConnection", "conn": conn}

	atomic.AddUint64(&poller.connectionCount, 1)

	rdr := bufio.NewReader(conn)

	for {
		conn.SetReadDeadline(time.Now().Add(poller.Timeout))
		line, err := rdr.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				LogError(ctx, err, "reading string")
			}
			break
		}

		measurement, err := poller.parseLine(line)
		if err != nil {
			atomic.AddUint64(&poller.parseErrorCount, 1)
			LogError(ctx, err, "parse error")
			break
		}

		poller.measurements <- measurement
		atomic.AddUint64(&poller.metricCount, 1)
	}
}

func (poller Listen) parseLine(line string) (Measurement, error) {
	var err error
	var when time.Time
	var value interface{}

	mType := "g"
	unit := Empty

	fields := strings.Fields(line)
	flen := len(fields)
	if flen < 3 || flen > 5 {
		return nil, fmt.Errorf("Expected 3, 4 or 5 fields, found %d", len(fields))
	}

	when, err = poller.parseDate(fields[0])
	if err != nil {
		return nil, err
	}

	if !MetricNameRegexp.MatchString(fields[1]) {
		return nil, fmt.Errorf("%q is an improper metric name", fields[1])
	}

	value, mType, err = poller.parseValue(fields[2])
	if err != nil {
		return nil, err
	}

	if flen >= 4 {
		if fields[3] == "c" || fields[3] == "counter" {
			switch value.(type) {
			case uint64:
				mType = "c"
			default:
				return nil, fmt.Errorf("value given is incompatible with counter type")
			}
		} else if fields[3] == "g" || fields[3] == "gauge" {
			mType = "g"
		} else {
			return nil, fmt.Errorf("type specified, but wasn't counter or gauge")
		}
	}

	if flen >= 5 {
		subs := UnitRegexp.FindStringSubmatch(fields[4])
		if len(subs) == 4 {
			unit = Unit{subs[1], subs[3]}
		} else {
			return nil, fmt.Errorf("invalid unit specified in: %q", fields[4])
		}
	}

	if mType == "c" {
		return CounterMeasurement{when, poller.Name(), strings.Fields(fields[1]), value.(uint64), unit}, nil
	}

	switch value.(type) {
	case float64:
		return FloatGaugeMeasurement{when, poller.Name(), strings.Fields(fields[1]), value.(float64), unit}, nil
	case uint64:
		return GaugeMeasurement{when, poller.Name(), strings.Fields(fields[1]), value.(uint64), unit}, nil
	default:
		return nil, fmt.Errorf("couldn't create gauge measurement")
	}
}

// Parse a string containing either an RFC3339 timestamp or a unix epoch timestamp
func (poller Listen) parseDate(ds string) (time.Time, error) {
	if when, err := time.Parse(time.RFC3339, ds); err == nil {
		return when, nil
	}

	// unix ts?
	if ts, terr := strconv.ParseUint(ds, 10, 64); terr != nil {
		return time.Now(), fmt.Errorf("Invalid timestamp: %q", ds)
	} else {
		return time.Unix(int64(ts), 0), nil
	}
}

func (poller Listen) parseValue(vs string) (interface{}, string, error) {
	val, err := strconv.ParseUint(vs, 10, 64)
	if err != nil {
		fval, err := strconv.ParseFloat(vs, 64)
		if err != nil {
			return nil, "", fmt.Errorf("Couldn't parse %q as value", vs)
		}
		return fval, "g", nil
	}

	return val, "c", nil
}
