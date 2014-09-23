package main

import (
	"bufio"
	"net"
	"time"
)

const (
	STATS_COMMAND = "STATS\n"
	HEADER        = "SourceName;SourceId;SourceInstance;State;Type;Number"
	FOOTER        = "."
)

type SyslogngStats struct {
	measurements chan<- Measurement
	Socket       string
}

func NewSyslogngStatsPoller(measurements chan<- Measurement, config Config) SyslogngStats {
	return SyslogngStats{
		measurements: measurements,
		Socket:       config.SyslogngSocket,
	}
}

func (poller SyslogngStats) Poll(tick time.Time) {
	conn, err := net.Dial("unix", poller.Socket)
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
				fields := Fields(line)
				poller.measurements <- CounterMeasurement{tick, poller.Name(), fields[:len(fields)-1], Atouint64(fields[len(fields)-1])}
			}
		}
	}
	conn.Close()
}

func (poller SyslogngStats) Name() string {
	return "syslog-ng-stats"
}

func (poller SyslogngStats) Exit() {}
