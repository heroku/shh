package pollers

import (
	"bufio"
	"github.com/freeformz/shh/mm"
	"github.com/freeformz/shh/utils"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
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

type Cpu struct {
	measurements chan<- *mm.Measurement
	last         map[string]CpuValues
}

func NewCpuPoller(measurements chan<- *mm.Measurement) Cpu {
	return Cpu{measurements: measurements, last: make(map[string]CpuValues)}
}

func calcPercent(val, total float64) string {
	if total == 0 {
		return "0"
	}
	return strconv.FormatFloat(val/total*100, 'f', 2, 64)
}

func (poller Cpu) Poll(tick time.Time) {
  var current, last CpuValues

	file, err := os.Open("/proc/stat")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		if strings.HasPrefix(line, "cpu") {
			fields := strings.Fields(line)
			cpu := fields[0]

			current = CpuValues{}

			current.User = utils.Atofloat64(fields[1])
			current.Nice = utils.Atofloat64(fields[2])
			current.System = utils.Atofloat64(fields[3])
			current.Idle = utils.Atofloat64(fields[4])
			current.Iowait = utils.Atofloat64(fields[5])
			current.Irq = utils.Atofloat64(fields[6])
			current.Softirq = utils.Atofloat64(fields[7])
			current.Steal = utils.Atofloat64(fields[8])
			current.Guest = utils.Atofloat64(fields[9])

			last = poller.last[cpu]

			if last.Total() != 0 {
				cTotal := current.Total() - last.Total()
				cUser := current.User - last.User
				cNice := current.Nice - last.Nice
				cSystem := current.System - last.System
				cIdle := current.Idle - last.Idle
				cIowait := current.Iowait - last.Iowait
				cIrq := current.Irq - last.Irq
				cSoftirq := current.Softirq - last.Softirq
				cSteal := current.Steal - last.Steal
				cGuest := current.Guest - last.Guest

				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "user"}, calcPercent(cUser, cTotal), mm.GAUGE}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "nice"}, calcPercent(cNice, cTotal), mm.GAUGE}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "system"}, calcPercent(cSystem, cTotal), mm.GAUGE}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "idle"}, calcPercent(cIdle, cTotal), mm.GAUGE}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "iowait"}, calcPercent(cIowait, cTotal), mm.GAUGE}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "irq"}, calcPercent(cIrq, cTotal), mm.GAUGE}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "softirq"}, calcPercent(cSoftirq, cTotal), mm.GAUGE}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "steal"}, calcPercent(cSteal, cTotal), mm.GAUGE}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "guest"}, calcPercent(cGuest, cTotal), mm.GAUGE}
			}

			poller.last[cpu] = current

		}
	}
}

func (poller Cpu) Name() string {
	return "cpu"
}
