package pollers

import (
	"bufio"
	"github.com/freeformz/shh/mm"
	"github.com/freeformz/shh/utils"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ListenStats struct {
	sync.RWMutex
	connectionCount float64
}

func (ls *ListenStats) IncrementConnectionCount() {
	ls.Lock()
	defer ls.Unlock()
	ls.connectionCount++
}

func (ls *ListenStats) DecrementConnectionCount() {
	ls.Lock()
	defer ls.Unlock()
	ls.connectionCount--
}

func (ls *ListenStats) ConnectionCount() float64 {
	ls.RLock()
	defer ls.RUnlock()
	return ls.connectionCount
}

type Listen struct {
	measurements chan<- *mm.Measurement
	listener     net.Listener
	stats        *ListenStats
}

const (
	DEFAULT_INTERVAL = "10s" // Default tick interval for pollers
)

var (
	listen      = utils.GetEnvWithDefault("SHH_LISTEN", "unix,/tmp/shh")
	interval    = utils.GetEnvWithDefaultDuration("SHH_INTERVAL", DEFAULT_INTERVAL)
	listenNet   string
	listenLaddr string
)

func init() {
	tmp := strings.Split(listen, ",")

	if len(tmp) != 2 {
		log.Fatal("SHH_LISTEN is not in the format: 'unix,/tmp/shh'")
	}

	listenNet = tmp[0]
	listenLaddr = tmp[1]

	switch listenNet {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
		break
	default:
		log.Fatalf("SHH_LISTEN format (%s,%s) is not correct", listenNet, listenLaddr)
	}

}

func NewListenPoller(measurements chan<- *mm.Measurement) Listen {
	listener, err := net.Listen(listenNet, listenLaddr)

	if err != nil {
		log.Fatal(err)
	}

	poller := Listen{measurements: measurements, listener: listener, stats: &ListenStats{}}

	go func(poller *Listen) {
		for {
			conn, err := poller.listener.Accept()
			if err != nil {
				log.Println(err)
				continue
			}

			go handleListenConnection(poller, conn)
		}
	}(&poller)

	return poller
}

func (poller Listen) Poll(tick time.Time) {
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"connection", "count"}, poller.stats.ConnectionCount()}
}

func handleListenConnection(poller *Listen, conn net.Conn) {
	defer conn.Close()

	var value interface{}

	poller.stats.IncrementConnectionCount()
	defer poller.stats.DecrementConnectionCount()

	r := bufio.NewReader(conn)

	for {
		conn.SetDeadline(time.Now().Add(interval).Add(interval))
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}

		fields := strings.Fields(line)
		if len(fields) == 3 {
			when, err := time.Parse(time.RFC3339, fields[0])
			if err != nil {
				break
			}
			value, err = strconv.ParseUint(fields[2], 10, 64)
			if err != nil {
				value, err = strconv.ParseFloat(fields[2], 64)
				if err != nil {
					break
				}
			}

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
