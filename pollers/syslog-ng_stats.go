package pollers

import (
	"github.com/freeformz/shh/config"
	"github.com/freeformz/shh/mm"
	"github.com/freeformz/shh/utils"
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
				fields := utils.Fields(line)
				poller.measurements <- &mm.Measurement{tick, poller.Name(), fields[:len(fields)-1], utils.Atouint64(fields[len(fields)-1])}
			}
		}
	}
	conn.Close()
}

func (poller SyslogngStats) Name() string {
	return "syslog-ng-stats"
}

func (poller SyslogngStats) Exit() {}
