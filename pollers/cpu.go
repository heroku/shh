package pollers

import (
	"github.com/freeformz/shh/mm"
	"github.com/freeformz/shh/utils"
	"strings"
	"time"
)

const (
	CPU_DATA = "/proc/stat"
)

type CpuValues struct {
	User    float64
	Nice    float64
	System  float64
	Idle    float64
	Iowait  float64
	Irq     float64
	Softirq float64
	Steal   float64
	Guest   float64
}

func (cv CpuValues) Total() float64 {
	return cv.User + cv.Nice + cv.System + cv.Idle + cv.Iowait + cv.Irq + cv.Softirq + cv.Steal + cv.Guest
}

func (cv CpuValues) DiffPercent(last CpuValues) CpuValues {
	totalDifference := cv.Total() - last.Total()
	if totalDifference == 0 {
		return CpuValues{}
	}
	return CpuValues{
		User:    (cv.User - last.User) / totalDifference * 100,
		Nice:    (cv.Nice - last.Nice) / totalDifference * 100,
		System:  (cv.System - last.System) / totalDifference * 100,
		Idle:    (cv.Idle - last.Idle) / totalDifference * 100,
		Iowait:  (cv.Iowait - last.Iowait) / totalDifference * 100,
		Irq:     (cv.Irq - last.Irq) / totalDifference * 100,
		Softirq: (cv.Softirq - last.Softirq) / totalDifference * 100,
		Steal:   (cv.Steal - last.Steal) / totalDifference * 100,
		Guest:   (cv.Guest - last.Guest) / totalDifference * 100,
	}
}

type Cpu struct {
	measurements chan<- *mm.Measurement
	last         map[string]CpuValues
}

func NewCpuPoller(measurements chan<- *mm.Measurement) Cpu {
	return Cpu{measurements: measurements, last: make(map[string]CpuValues)}
}

func (poller Cpu) Poll(tick time.Time) {
	var current, percent CpuValues

	for line := range utils.FileLineChannel(CPU_DATA) {
		if strings.HasPrefix(line, "cpu") {
			fields := strings.Fields(line)
			cpu := fields[0]

			current = CpuValues{
				User:    utils.Atofloat64(fields[1]),
				Nice:    utils.Atofloat64(fields[2]),
				System:  utils.Atofloat64(fields[3]),
				Idle:    utils.Atofloat64(fields[4]),
				Iowait:  utils.Atofloat64(fields[5]),
				Irq:     utils.Atofloat64(fields[6]),
				Softirq: utils.Atofloat64(fields[7]),
				Steal:   utils.Atofloat64(fields[8]),
				Guest:   utils.Atofloat64(fields[9]),
			}

			last, exists := poller.last[cpu]

			if exists {
				percent = current.DiffPercent(last)

				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "user"}, percent.User}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "nice"}, percent.Nice}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "system"}, percent.System}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "idle"}, percent.Idle}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "iowait"}, percent.Iowait}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "irq"}, percent.Irq}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "softirq"}, percent.Softirq}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "steal"}, percent.Steal}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "guest"}, percent.Guest}
			}

			poller.last[cpu] = current

		}
	}
}

func (poller Cpu) Name() string {
	return "cpu"
}

func (poller Cpu) Exit() {}
