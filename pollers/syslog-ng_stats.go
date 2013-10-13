package pollers

import (
	"fmt"
	"github.com/freeformz/shh/config"
	"github.com/freeformz/shh/mm"
	"net"
	"time"
	"bufio"
)

const (
	STATS_COMMAND = "STATS\n"
	HEADER = "SourceName;SourceId;SourceInstance;State;Type;Number"
	FOOTER = "."
)

type SyslogngStats struct {
	measurements chan<- *mm.Measurement
}

func NewSyslogngStatsPoller(measurements chan<- *mm.Measurement) SyslogngStats {
	return SyslogngStats{measurements: measurements}
}

func (poller SyslogngStats) Poll(tick time.Time) {
	conn, err := net.Dial("unix", config.SyslogngSocket)
	if err != nil {
		panic(err)
	}
	conn.Write([]byte(STATS_COMMAND))
  scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		if line == FOOTER {
			break
		} else {
			if line == HEADER {
				continue
			} else {
				fmt.Println(scanner.Text())
			}
		}
	}
	conn.Close()
}

func (poller SyslogngStats) Name() string {
	return "syslog-ng-stats"
}

func (poller SyslogngStats) Exit() {}
