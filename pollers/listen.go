package pollers

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
	"github.com/freeformz/shh/config"
	"github.com/freeformz/shh/mm"
	"github.com/freeformz/shh/utils"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	listenNet   string
	listenLaddr string
)

func init() {
	ctx := utils.Slog{"poller": "listen", "fn": "init"}
	tmp := strings.Split(config.Listen, ",")

	formatErr := fmt.Errorf("SHH_LISTEN is not in the correct format")
	fix := "The correct format is: <tcp|tcp4|tcp6|unix|unixpacket>,<address>"

	if len(tmp) != 2 {
		ctx.FatalError(formatErr, fix)
	}

	listenNet = tmp[0]
	listenLaddr = tmp[1]

	switch listenNet {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
		break
	default:
		ctx.FatalError(formatErr, fix)
	}

	// If this is a path, remove it
	if listenNet == "unix" && utils.Exists(listenLaddr) {
		err := os.Remove(listenLaddr)
		if err != nil {
			ctx.FatalError(err, "unable to remove old socket path")
		}
	}

}

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
	measurements chan<- *mm.Measurement
	listener     net.Listener
	stats        *ListenStats
}

func NewListenPoller(measurements chan<- *mm.Measurement) Listen {
	ctx := utils.Slog{"poller": "listen", "fn": "NewListenPoller"}
	listener, err := net.Listen(listenNet, listenLaddr)

	if err != nil {
		ctx.FatalError(err, "unable to listen on "+listenNet+listenLaddr)
	}

	ls := &ListenStats{counts: make(map[string]interface{})}
	ls.New("connection.count", 0.0)
	ls.New("time.parse.errors", 0)
	ls.New("value.parse.errors", 0)
	ls.New("metrics", 0)

	poller := Listen{measurements: measurements, listener: listener, stats: ls}

	go func(poller *Listen) {
		ctx := utils.Slog{"poller": poller.Name(), "fn": "acceptor"}

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
		poller.measurements <- &mm.Measurement{tick, poller.Name(), strings.Split("stats."+k, "."), poller.stats.CountOf(k)}
	}
}

func handleListenConnection(poller *Listen, conn net.Conn) {
	defer conn.Close()

	ctx := utils.Slog{"poller": poller.Name(), "fn": "handleListenConnection", "conn": conn}

	var value interface{}

	poller.stats.Increment("connection.count")
	defer poller.stats.Decrement("connection.count")

	r := bufio.NewReader(conn)

	for {
		conn.SetDeadline(time.Now().Add(config.Interval).Add(config.Interval))
		line, err := r.ReadString('\n')
		if err != nil {
			ctx.Error(err, "reading string")
			break
		}

		fields := strings.Fields(line)
		if len(fields) == 3 {
			when, err := time.Parse(time.RFC3339, fields[0])
			if err != nil {
				ctx.Error(err, "parsing time")
				poller.stats.Increment("time.parse.errors")
				break
			}
			value, err = strconv.ParseUint(fields[2], 10, 64)
			if err != nil {
				value, err = strconv.ParseFloat(fields[2], 64)
				if err != nil {
					ctx.Error(err, "parsing float / int")
					poller.stats.Increment("value.parse.errors")
					break
				}
			}

			poller.stats.Increment("metrics")

			poller.measurements <- &mm.Measurement{when, poller.Name(), strings.Fields(fields[1]), value}
		}
	}
}

func (poller Listen) Name() string {
	return "listen"
}

func (poller Listen) Exit() {
	poller.listener.Close()
}
