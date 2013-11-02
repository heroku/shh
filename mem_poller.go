package shh

import (
	"time"
)

const (
	MEMORY_FILE = "/proc/meminfo"
)

type Memory struct {
	measurements chan<- *Measurement
}

func NewMemoryPoller(measurements chan<- *Measurement) Memory {
	return Memory{measurements: measurements}
}

// http://www.kernel.org/doc/Documentation/filesystems/proc.txt
func (poller Memory) Poll(tick time.Time) {

	for line := range FileLineChannel(MEMORY_FILE) {
		fields := Fields(line)
		fixed_names := FixUpName(fields[0])
		value := Atofloat64(fields[1])
		if len(fields) == 3 && fields[2] == "kB" {
			value = value * 1024.0
		}
		poller.measurements <- &Measurement{tick, poller.Name(), fixed_names, value}
	}
}

func (poller Memory) Name() string {
	return "mem"
}

func (poller Memory) Exit() {}
