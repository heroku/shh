package main

import (
	"time"
)

const (
	MEMORY_FILE = "/proc/meminfo"
)

type Memory struct {
	measurements   chan<- Measurement
	memPercentage  bool
	swapPercentage bool
}

func NewMemoryPoller(measurements chan<- Measurement, config Config) Memory {
	memPerc := LinearSliceContainsString(config.Percentages, "mem")
	swapPerc := LinearSliceContainsString(config.Percentages, "swap")
	return Memory{measurements: measurements, memPercentage: memPerc, swapPercentage: swapPerc}
}

// http://www.kernel.org/doc/Documentation/filesystems/proc.txt
func (poller Memory) Poll(tick time.Time) {
	unit := Empty
	memTotal := uint64(0)
	memFree := uint64(0)
	swapTotal := uint64(0)
	swapFree := uint64(0)

	for line := range FileLineChannel(MEMORY_FILE) {
		fields := Fields(line)
		fixed_names := FixUpName(fields[0])
		value := Atouint64(fields[1])
		if len(fields) == 3 && fields[2] == "kB" {
			value = value * 1024.0
			unit = Bytes
		} else {
			unit = Empty
		}

		poller.measurements <- GaugeMeasurement{tick, poller.Name(), fixed_names, value, unit}

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

	if poller.memPercentage && memTotal > 0 && memFree >= 0 {
		poller.measurements <- FloatGaugeMeasurement{tick, poller.Name(),
			[]string{"memtotal", "perc"}, 100.0 * float64(memTotal - memFree) / float64(memTotal), Percent}
	}

	if poller.swapPercentage && swapTotal > 0.0 && swapFree >= 0.0 {
		poller.measurements <- FloatGaugeMeasurement{tick, poller.Name(),
			[]string{"swaptotal", "perc"}, 100.0 * float64(swapTotal - swapFree) / float64(swapTotal), Percent}
	}
}

func (poller Memory) Name() string {
	return "mem"
}

func (poller Memory) Exit() {}
