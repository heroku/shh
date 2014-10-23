package shh

import (
	"time"
)

const (
	MEMORY_FILE = "/proc/meminfo"
)

var (
	MEM_MINIMAL_LIST = []string{"memfree", "memtotal", "swapfree", "swaptotal", "buffers", "cached", "swapcached"}
)

type Memory struct {
	measurements   chan<- Measurement
	memPercentage  bool
	swapPercentage bool
	full           bool
}

func NewMemoryPoller(measurements chan<- Measurement, config Config) Memory {
	memPerc := LinearSliceContainsString(config.Percentages, "mem")
	swapPerc := LinearSliceContainsString(config.Percentages, "swap")
	mem := Memory{
		measurements:   measurements,
		memPercentage:  memPerc,
		swapPercentage: swapPerc,
	}
	mem.full = SliceContainsString(config.Full, mem.Name())
	return mem
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

		if !poller.full && !SliceContainsString(MEM_MINIMAL_LIST, fixed_names[0]) {
			continue
		}

		poller.measurements <- GaugeMeasurement{tick, poller.Name(), fixed_names, value, unit}
	}

	if poller.memPercentage && memTotal > 0 && memFree >= 0 {
		poller.measurements <- FloatGaugeMeasurement{tick, poller.Name(),
			[]string{"memtotal", "perc"}, 100.0 * float64(memTotal-memFree) / float64(memTotal), Percent}
	}

	if poller.swapPercentage && swapTotal > 0.0 && swapFree >= 0.0 {
		poller.measurements <- FloatGaugeMeasurement{tick, poller.Name(),
			[]string{"swaptotal", "perc"}, 100.0 * float64(swapTotal-swapFree) / float64(swapTotal), Percent}
	}

}

func (poller Memory) Name() string {
	return "mem"
}

func (poller Memory) Exit() {}
