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
)

var (
	MetricNameRegexp = regexp.MustCompile("^[a-zA-Z0-9]([a-zA-Z0-9.-]+)?$")
	MetaRegexp       = regexp.MustCompile("^([cg])(:([a-zA-Z$%#]+)(,([a-zA-Z$%#]+))?)?$") // <type 1>:<unit 3>,<abbr 5>
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
	ctx := Slog{"poller": "listen", "fn": "NewListenPoller"}
	tmp := strings.Split(config.Listen, ",")

	formatErr := fmt.Errorf("SHH_LISTEN is not in the correct format")
	fix := "The correct format is: <tcp|tcp4|tcp6|unix|unixpacket>,<address>"

	if len(tmp) != 2 {
		ctx.FatalError(formatErr, fix)
	}

	listenNet := tmp[0]
	listenLaddr := tmp[1]

	switch listenNet {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
		break
	default:
		ctx.FatalError(formatErr, fix)
	}

	// If this is a path, remove it
	if listenNet == "unix" && Exists(listenLaddr) {
		err := os.Remove(listenLaddr)
		if err != nil {
			ctx.FatalError(err, "unable to remove old socket path")
		}
	}

	listener, err := net.Listen(listenNet, listenLaddr)

	if err != nil {
		ctx.FatalError(err, "unable to listen on "+listenNet+listenLaddr)
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
		ctx := Slog{"poller": poller.Name(), "fn": "acceptor"}

		for {
			conn, err := poller.listener.Accept()
			if err != nil {
				ctx.Error(err, "accepting connection")
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

	ctx := Slog{"poller": poller.Name(), "fn": "handleListenConnection", "conn": conn}

	poller.stats.Increment("connection.count")
	defer poller.stats.Decrement("connection.count")

	r := bufio.NewReader(conn)

	for {
		conn.SetDeadline(time.Now().Add(poller.Interval))
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			ctx.Error(err, "reading string")
			break
		}

		measurement, err := poller.parseLine(line)
		if err != nil {
			ctx.Error(err, "parse error")
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
	fields := strings.Fields(line)
	if len(fields) != 3 {
		return nil, fmt.Errorf("Expected 3 fields, found %d", len(fields))
	}

	when, err := poller.parseDate(fields[0])
	if err != nil {
		return nil, fmt.Errorf("unable to parse date: %q", fields[0])
	}

	if !MetricNameRegexp.MatchString(fields[1]) {
		return nil, fmt.Errorf("%q is an improper metric name", fields[1])
	}

	return poller.parseMeasurement(when, fields[1], fields[2])
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

func (poller Listen) parseMeasurement(when time.Time, metric string, sval string) (Measurement, error) {
	bits := strings.Split(sval, "|") // value, meta
	if len(bits) == 2 {
		subs := MetaRegexp.FindStringSubmatch(bits[1])
		if len(subs) == 0 {
			poller.stats.Increment("meta.parse.errors")
			return nil, fmt.Errorf("Couldn't parse %q as meta information", bits[1])
		}

		unit := Unit{subs[3], subs[5]}
		switch subs[1] {
		case "c":
			val, err := strconv.ParseUint(bits[0], 10, 64)
			if err != nil {
				poller.stats.Increment("value.parse.errors")
				return nil, fmt.Errorf("Couldn't parse %q as counter", bits[0])
			}
			return CounterMeasurement{when, poller.Name(), strings.Fields(metric), val, unit}, nil
		case "g":
			val, err := strconv.ParseFloat(bits[0], 64)
			if err != nil {
				poller.stats.Increment("value.parse.errors")
				return nil, fmt.Errorf("Couldn't parse %q as gauge", bits[0])
			}
			return FloatGaugeMeasurement{when, poller.Name(), strings.Fields(metric), val, unit}, nil
		default:
			return nil, fmt.Errorf("Unable to determine measurement type")
		}
	} else if len(bits) == 1 { // just a value, infer type
		val, err := strconv.ParseUint(bits[0], 10, 64)
		if err != nil {
			fval, err := strconv.ParseFloat(bits[0], 64)
			if err != nil {
				poller.stats.Increment("value.parse.errors")
				return nil, fmt.Errorf("Couldn't parse %q as value", bits[0])
			}
			return FloatGaugeMeasurement{when, poller.Name(), strings.Fields(metric), fval, Empty}, nil
		}

		return CounterMeasurement{when, poller.Name(), strings.Fields(metric), val, Empty}, nil
	}

	poller.stats.Increment("value.parse.errors")
	return nil, fmt.Errorf("Invalid value")
}
