package pollers

import (
	"bufio"
	"github.com/freeformz/shh/mm"
	"log"
	"os"
	"strings"
	"time"
)

type Load struct {
	measurements chan<- *mm.Measurement
}

func NewLoadPoller(measurements chan<- *mm.Measurement) Load {
	return Load{measurements: measurements}
}

func (poller Load) Poll(tick time.Time) {
	file, err := os.Open("/proc/loadavg")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	line, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	fields := strings.Fields(line)
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"1m"}, fields[0], mm.GAUGE}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"5m"}, fields[1], mm.GAUGE}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"15m"}, fields[2], mm.GAUGE}
	entities := strings.Split(fields[3], "/")
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"scheduling", "entities", "executing"}, entities[0], mm.GAUGE}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"scheduling", "entities", "total"}, entities[1], mm.GAUGE}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"pid", "last"}, fields[4], mm.GAUGE}
}

func (poller Load) Name() string {
	return "load"
}
