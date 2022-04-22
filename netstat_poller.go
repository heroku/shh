package shh

import (
	"syscall"
	"time"

	"github.com/heroku/slog"
	"github.com/shirou/gopsutil/net"
)

type Netstat struct {
	measurements chan<- Measurement
}

func NewNetstatPoller(measurements chan<- Measurement) Netstat {
	return Netstat{measurements: measurements}
}

func getStats() (map[string]int, error) {
	conns, err := net.Connections("all")
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	counts["UDP"] = 0
	for _, conn := range conns {
		if conn.Type == syscall.SOCK_DGRAM {
			counts["UDP"]++
			continue
		}
		c, ok := counts[conn.Status]
		if !ok {
			counts[conn.Status] = 0
		}
		counts[conn.Status] = c + 1
	}

	fields := map[string]int{
		"tcp_established": counts["ESTABLISHED"],
		"tcp_syn_sent":    counts["SYN_SENT"],
		"tcp_syn_recv":    counts["SYN_RECV"],
		"tcp_fin_wait1":   counts["FIN_WAIT1"],
		"tcp_fin_wait2":   counts["FIN_WAIT2"],
		"tcp_time_wait":   counts["TIME_WAIT"],
		"tcp_close":       counts["CLOSE"],
		"tcp_close_wait":  counts["CLOSE_WAIT"],
		"tcp_last_ack":    counts["LAST_ACK"],
		"tcp_listen":      counts["LISTEN"],
		"tcp_closing":     counts["CLOSING"],
		"tcp_none":        counts["NONE"],
		"udp_socket":      counts["UDP"],
	}

	return fields, nil
}

func (poller Netstat) Poll(tick time.Time) {
	ctx := slog.Context{"poller": poller.Name(), "fn": "Poll", "tick": tick}

	stats, err := getStats()
	if err != nil {
		LogError(ctx, err, "Error reading netconn")
		return
	}

	for field, _ := range stats {
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{field}, uint64(stats[field]), Connections}
	}
}

func (poller Netstat) Name() string {
	return "netstat"
}

func (poller Netstat) Exit() {}
