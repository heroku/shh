package main

import (
	"strings"
	"time"
)

const (
	NETSTAT_FILE = "/proc/net/netstat"
)

type Netstat struct {
	measurements chan<- Measurement
	includedProtos    []string
	includedStats     []string
}

func NewNetstatPoller(measurements chan<- Measurement, config Config) Netstat {
	return Netstat{
		measurements: measurements,
	  includedProtos: config.NetstatProtos,
		includedStats:     config.NetstatStats,
	}
}

func (poller Netstat) Poll(tick time.Time) {
	var proto string
	var fields, values []string

	li := 0
	for line := range FileLineChannel(NETSTAT_FILE) {
		if li%2 == 1 {
			values = Fields(line)

			values[0] = strings.Replace(values[0], ":", "", -1)
			if len(poller.includedProtos) == 0 || LinearSliceContainsString(poller.includedProtos, values[0]) {
				proto = strings.ToLower(values[0])
				for i := 1; i < len(values); i++ {
					if len(poller.includedStats) == 0 || LinearSliceContainsString(poller.includedStats, fields[i]) {
						poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{proto, strings.ToLower(fields[i])}, Atouint64(values[i])}
					}
				}
			}
		} else {
			fields = Fields(line)
		}
		li += 1
	}
}

func (poller Netstat) Name() string {
	return "netstat"
}

func (poller Netstat) Exit() {}
