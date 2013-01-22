package pollers

import (
	"bufio"
	"github.com/freeformz/shh/mm"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

type Cpu struct {
	measurements chan<- *mm.Measurement
}

func NewCpuPoller(measurements chan<- *mm.Measurement) Cpu {
	return Cpu{measurements: measurements}
}

func (poller Cpu) Poll(tick time.Time) {
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
			poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "user"}, fields[1], mm.COUNTER}
			poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "nice"}, fields[2], mm.COUNTER}
			poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "system"}, fields[3], mm.COUNTER}
			poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "idle"}, fields[4], mm.COUNTER}
			poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "iowait"}, fields[5], mm.COUNTER}
			poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "irq"}, fields[6], mm.COUNTER}
			poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "softirq"}, fields[7], mm.COUNTER}
			poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "steal"}, fields[8], mm.COUNTER}
			poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{cpu, "guest"}, fields[9], mm.COUNTER}
		}
	}
}

func (poller Cpu) Name() string {
	return "cpu"
}
