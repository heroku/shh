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
	"sync"
	"time"

	"github.com/heroku/slog"
)

var (
	MetricNameRegexp = regexp.MustCompile("^[a-zA-Z0-9]([a-zA-Z0-9.-]+)?$")
	UnitRegexp       = regexp.MustCompile("^([a-zA-Z$%#]+)(,([a-zA-Z$%#]+))?$") // <unit 1>,<abbr 3>
)

// Used to track global listen stats
type ListenStats struct {
	sync.RWMutex
	counts map[string]interface{}
}

func (ls *ListenStats) New(what string, initialValue interface{}) {
	ls.Lock()
	defer ls.Unlock()
	switch initialValue.(type) {
	case float64, uint64:
		ls.counts[what] = initialValue
	case int:
		ls.counts[what] = uint64(initialValue.(int))
	}
}

func (ls *ListenStats) Increment(what string) {
	ls.Lock()
	defer ls.Unlock()
	v := ls.counts[what]
	switch v.(type) {
	case float64:
		tmp := v.(float64)
		tmp++
		ls.counts[what] = tmp
	case uint64, int:
		tmp := v.(uint64)
		tmp++
		ls.counts[what] = tmp
	}
}

func (ls *ListenStats) Decrement(what string) {
	ls.Lock()
	defer ls.Unlock()
	v := ls.counts[what]
	switch v.(type) {
	case float64:
		tmp := v.(float64)
		tmp--
		ls.counts[what] = tmp
	case uint64, int:
		tmp := v.(uint64)
		tmp--
		ls.counts[what] = tmp
	}
}

func (ls *ListenStats) CountOf(what string) interface{} {
	ls.RLock()
	defer ls.RUnlock()
	return ls.counts[what]
}

func (ls *ListenStats) Keys() <-chan string {
	ls.RLock()

	c := make(chan string)

	go func(c chan<- string) {
		defer ls.RUnlock()
		defer close(c)
		for k, _ := range ls.counts {
			c <- k
		}
	}(c)

	return c
}

type Listen struct {
	measurements chan<- Measurement
	listener     net.Listener
	stats        *ListenStats
	Interval     time.Duration
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

	ls := &ListenStats{counts: make(map[string]interface{})}
	ls.New("connection.count", 0.0)
	ls.New("time.parse.errors", 0)
	ls.New("value.parse.errors", 0)
	ls.New("metrics", 0)

	poller := Listen{
		measurements: measurements,
		listener:     listener,
		stats:        ls,
		Interval:     config.Interval,
	}

	go func(poller *Listen) {
		ctx := slog.Context{"poller": poller.Name(), "fn": "acceptor"}

		for {
			conn, err := poller.listener.Accept()
			if err != nil {
				LogError(ctx, err, "accepting connection")
				continue
			}

			go handleListenConnection(poller, conn)
		}
	}(&poller)

	return poller
}

func (poller Listen) Poll(tick time.Time) {
	for k := range poller.stats.Keys() {
		v := poller.stats.CountOf(k)
		switch v.(type) {
		case float64:
			poller.measurements <- FloatGaugeMeasurement{tick, poller.Name(), strings.Split("stats."+k, "."), v.(float64), Empty}
		case uint64:
			poller.measurements <- CounterMeasurement{tick, poller.Name(), strings.Split("stats."+k, "."), v.(uint64), Empty}
		case int:
			poller.measurements <- CounterMeasurement{tick, poller.Name(), strings.Split("stats."+k, "."), uint64(v.(int)), Empty}
		}
	}
}

func handleListenConnection(poller *Listen, conn net.Conn) {
	defer conn.Close()

	ctx := slog.Context{"poller": poller.Name(), "fn": "handleListenConnection", "conn": conn}

	poller.stats.Increment("connection.count")
	defer poller.stats.Decrement("connection.count")

	r := bufio.NewReader(conn)

	for {
		conn.SetDeadline(time.Now().Add(poller.Interval))
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			LogError(ctx, err, "reading string")
			break
		}

		measurement, err := poller.parseLine(line)
		if err != nil {
			LogError(ctx, err, "parse error")
			break
		}

		poller.measurements <- measurement
		poller.stats.Increment("metrics")
	}
}

func (poller Listen) Name() string {
	return "listen"
}

func (poller Listen) Exit() {
	poller.listener.Close()
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
		return nil, fmt.Errorf("unable to parse date: %q", fields[0])
	}

	if !MetricNameRegexp.MatchString(fields[1]) {
		return nil, fmt.Errorf("%q is an improper metric name", fields[1])
	}

	value, mType, err = poller.parseValue(fields[2])
	if err != nil {
		return nil, fmt.Errorf("unable to parse value: %s", err)
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
			poller.stats.Increment("meta.parse.errors")
			return nil, fmt.Errorf("type specified, but wasn't counter or gauge")
		}
	}

	if flen >= 5 {
		subs := UnitRegexp.FindStringSubmatch(fields[4])
		if len(subs) == 4 {
			unit = Unit{subs[1], subs[3]}
		} else {
			poller.stats.Increment("meta.parse.errors")
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

func (poller Listen) parseDate(ds string) (time.Time, error) {
	if when, err := time.Parse(time.RFC3339, ds); err == nil {
		return when, nil
	}

	// unix ts?
	if ts, terr := strconv.ParseUint(ds, 10, 64); terr != nil {
		poller.stats.Increment("time.parse.errors")
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
			poller.stats.Increment("value.parse.errors")
			return nil, "", fmt.Errorf("Couldn't parse %q as value", vs)
		}
		return fval, "g", nil
	}

	return val, "c", nil
}
