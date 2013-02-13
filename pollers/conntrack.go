package pollers

import (
	"github.com/freeformz/filechan"
	"github.com/freeformz/shh/mm"
	"github.com/freeformz/shh/utils"
	"time"
)

const (
	CONNTRACK_DATA = "/proc/net/ip_conntrack_sdfsfd"
)

type Conntrack struct {
	measurements chan<- *mm.Measurement
}

func NewConntrackPoller(measurements chan<- *mm.Measurement) Conntrack {
	return Conntrack{measurements: measurements}
}

func (poller Conntrack) Poll(tick time.Time) {

	var udp, tcp float64
	var closed, listen, synrcvd, synsent, established, closewait, lastack, finwait1, finwait2, closing, timewait float64
	var unreplied, assured float64

	connInfo, err := filechan.FileLineChannel(CONNTRACK_DATA)

	if err != nil {
		return
	}

	for line := range connInfo {
		fields := utils.Fields(line)
		switch fields[0] {
		case "tcp":
			tcp++
			switch fields[3] {
			case "CLOSED":
				closed++
			case "LISTEN":
				listen++
			case "SYN_RCVD":
				synrcvd++
			case "SYN_SENT":
				synsent++
			case "ESTABLISHED":
				established++
			case "CLOSE_WAIT":
				closewait++
			case "LAST_ACK":
				lastack++
			case "FIN_WAIT_1":
				finwait1++
			case "FIN_WAIT_2":
				finwait2++
			case "CLOSING":
				closing++
			case "TIME_WAIT":
				timewait++
			}
		case "udp":
			udp++
		}
		switch {
		case utils.LinearSliceContainsString(fields, "[UNREPLIED]"):
			unreplied++
		case utils.LinearSliceContainsString(fields, "[ASSURED]"):
			assured++
		}
	}

	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"tcp", "all"}, tcp}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"tcp", "closed"}, closed}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"tcp", "listen"}, listen}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"tcp", "syn", "rcvd"}, synrcvd}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"tcp", "syn", "sent"}, synsent}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"tcp", "established"}, established}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"tcp", "closewait"}, closewait}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"tcp", "lastack"}, lastack}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"tcp", "fin", "wait", "1"}, finwait1}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"tcp", "fin", "wait", "2"}, finwait2}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"tcp", "closing"}, closing}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"tcp", "timewait"}, timewait}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"udp", "all"}, udp}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"unreplied"}, unreplied}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"assured"}, assured}
}

func (poller Conntrack) Name() string {
	return "conntrack"
}

func (poller Conntrack) Exit() {}
