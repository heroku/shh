package main

import (
	"time"
)

const (
	MEMORY_FILE = "/proc/meminfo"
)

type Memory struct {
	measurements chan<- *Measurement
	memPercentage bool
	swapPercentage bool
}

func NewMemoryPoller(measurements chan<- *Measurement, config Config) Memory {
	memPerc := LinearSliceContainsString(config.Percentages, "mem")
	swapPerc := LinearSliceContainsString(config.Percentages, "swap")
	return Memory{measurements: measurements, memPercentage: memPerc, swapPercentage: swapPerc}
}

// http://www.kernel.org/doc/Documentation/filesystems/proc.txt
func (poller Memory) Poll(tick time.Time) {
	memTotal := float64(-1)
	memFree := float64(-1)
	swapTotal := float64(-1)
	swapFree := float64(-1)

	for line := range FileLineChannel(MEMORY_FILE) {
		fields := Fields(line)
		fixed_names := FixUpName(fields[0])
		value := Atofloat64(fields[1])
		if len(fields) == 3 && fields[2] == "kB" {
			value = value * 1024.0
		}
		poller.measurements <- &Measurement{tick, poller.Name(), fixed_names, value}

		switch fixed_names[0] {
		case "memtotal":
			memTotal = value
		case "memfree":
			memFree = value
		case "swaptotal":
			swapTotal = value
		case "swapfree":
			swapFree = value
		}
	}

	if poller.memPercentage && memTotal > 0.0 && memFree >= 0.0 {
		poller.measurements <- &Measurement{tick, poller.Name(),
			[]string{"memtotal", "perc"}, (memTotal - memFree) / memTotal}
	}

	if poller.swapPercentage && swapTotal > 0.0 && swapFree >= 0.0 {
		poller.measurements <- &Measurement{tick, poller.Name(),
			[]string{"swaptotal", "perc"}, (swapTotal - swapFree) / swapTotal}
	}
}

func (poller Memory) Name() string {
	return "mem"
}

func (poller Memory) Exit() {}
