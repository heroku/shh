package pollers

import (
	"github.com/freeformz/shh/mm"
	"github.com/freeformz/shh/utils"
	"time"
)

const (
	MEMORY_FILE = "/proc/meminfo"
)

type Memory struct {
	measurements chan<- *mm.Measurement
}

func NewMemoryPoller(measurements chan<- *mm.Measurement) Memory {
	return Memory{measurements: measurements}
}

// http://www.kernel.org/doc/Documentation/filesystems/proc.txt
func (poller Memory) Poll(tick time.Time) {

	for line := range utils.FileLineChannel(MEMORY_FILE) {
		fields := utils.Fields(line)
		fixed_names := utils.FixUpName(fields[0])
		value := utils.Atofloat64(fields[1])
		if len(fields) == 3 && fields[2] == "kB" {
			value = value * 1024.0
		}
		poller.measurements <- &mm.Measurement{tick, poller.Name(), fixed_names, value}
	}
}

func (poller Memory) Name() string {
	return "mem"
}

func (poller Memory) Exit() {}
