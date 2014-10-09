package main

import (
	"strings"
	"time"
)

const (
	SOCKSTAT4 = "/proc/net/sockstat"
	SOCKSTAT6 = "/proc/net/sockstat6"
)

type SockStat struct {
	measurements chan<- Measurement
	files        []string
	Protocols    []string
}

func NewSockStatPoller(measurements chan<- Measurement, config Config) SockStat {
	var includeV6, includeV4 bool
	for _, proto := range config.SockStatProtos {
		if strings.HasSuffix(proto, "6") {
			includeV6 = true
		} else {
			includeV4 = true
		}
	}

	filesToPoll := make([]string, 0)
	if includeV4 {
		filesToPoll = append(filesToPoll, SOCKSTAT4)
	}
	if includeV6 {
		filesToPoll = append(filesToPoll, SOCKSTAT6)
	}

	return SockStat{
		measurements: measurements,
		files:        filesToPoll,
		Protocols:    config.SockStatProtos,
	}
}

// http://www.kernel.org/doc/Documentation/filesystems/proc.txt (section 1.4)
func (poller SockStat) Poll(tick time.Time) {
	for _, file := range poller.files {
		for line := range FileLineChannel(file) {
			fields := Fields(line)
			proto := fields[0]

			if SliceContainsString(poller.Protocols, proto) && len(fields) > 1 {
				for i := 1; i+1 < len(fields); i += 2 {
					unit := Sockets
					if fields[i] == "mem" {
						unit = Empty
					}
				poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{strings.ToLower(proto), fields[i]}, Atouint64(fields[i+1]), unit}
				}
			}
		}
	}
}

func (poller SockStat) Name() string {
	return "sockstat"
}

func (poller SockStat) Exit() {}
