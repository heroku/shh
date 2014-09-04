package main

import (
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
	measurements  chan<- Measurement
	AggregateOnly bool
	last          map[string]CpuValues
}

func NewCpuPoller(measurements chan<- Measurement, config Config) Cpu {
	return Cpu{
		measurements:  measurements,
		last:          make(map[string]CpuValues),
		AggregateOnly: config.CpuOnlyAggregate,
	}
}

func (poller Cpu) Poll(tick time.Time) {
	var current, percent CpuValues

	for line := range FileLineChannel(CPU_DATA) {
		if strings.HasPrefix(line, "cpu") {
			fields := strings.Fields(line)
			cpu := fields[0]

			if poller.AggregateOnly && cpu != "cpu" {
				continue
			}

			current = CpuValues{
				User:    Atofloat64(fields[1]),
				Nice:    Atofloat64(fields[2]),
				System:  Atofloat64(fields[3]),
				Idle:    Atofloat64(fields[4]),
				Iowait:  Atofloat64(fields[5]),
				Irq:     Atofloat64(fields[6]),
				Softirq: Atofloat64(fields[7]),
				Steal:   Atofloat64(fields[8]),
			}

			if len(fields) > 9 {
				current.Guest = Atofloat64(fields[9])
			} else {
				current.Guest = 0
			}

			last, exists := poller.last[cpu]

			if exists {
				percent = current.DiffPercent(last)

				poller.measurements <- &FloatGaugeMeasurement{tick, poller.Name(), []string{cpu, "user"}, percent.User}
				poller.measurements <- &FloatGaugeMeasurement{tick, poller.Name(), []string{cpu, "nice"}, percent.Nice}
				poller.measurements <- &FloatGaugeMeasurement{tick, poller.Name(), []string{cpu, "system"}, percent.System}
				poller.measurements <- &FloatGaugeMeasurement{tick, poller.Name(), []string{cpu, "idle"}, percent.Idle}
				poller.measurements <- &FloatGaugeMeasurement{tick, poller.Name(), []string{cpu, "iowait"}, percent.Iowait}
				poller.measurements <- &FloatGaugeMeasurement{tick, poller.Name(), []string{cpu, "irq"}, percent.Irq}
				poller.measurements <- &FloatGaugeMeasurement{tick, poller.Name(), []string{cpu, "softirq"}, percent.Softirq}
				poller.measurements <- &FloatGaugeMeasurement{tick, poller.Name(), []string{cpu, "steal"}, percent.Steal}
				poller.measurements <- &FloatGaugeMeasurement{tick, poller.Name(), []string{cpu, "guest"}, percent.Guest}
			}

			poller.last[cpu] = current

		}
	}
}

func (poller Cpu) Name() string {
	return "cpu"
}

func (poller Cpu) Exit() {}
