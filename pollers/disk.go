package pollers

import (
	"bufio"
	"github.com/freeformz/shh/mm"
	"github.com/freeformz/shh/utils"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

const (
	SYS = "/sys/block/"
)

type Disk struct {
	measurements chan<- *mm.Measurement
}

func NewDiskPoller(measurements chan<- *mm.Measurement) Disk {
	return Disk{measurements: measurements}
}

// http://www.kernel.org/doc/Documentation/block/stat.txt
func (poller Disk) Poll(tick time.Time) {
	devices := make(chan string)
	go feedDevices(devices)

	for device := range devices {
		statBytes, err := ioutil.ReadFile(SYS + device + "/stat")
		if err != nil {
			log.Fatal(err)
		}

		fields := strings.Fields(string(statBytes))
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{device, "read", "requests"}, fields[0], mm.COUNTER}
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{device, "read", "merges"}, fields[1], mm.COUNTER}
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{device, "read", "sectors"}, fields[2], mm.COUNTER}
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{device, "read", "ticks"}, fields[3], mm.COUNTER}

		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{device, "write", "requests"}, fields[4], mm.COUNTER}
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{device, "write", "merges"}, fields[5], mm.COUNTER}
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{device, "write", "sectors"}, fields[6], mm.COUNTER}
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{device, "write", "ticks"}, fields[7], mm.COUNTER}

		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{device, "in_flight", "requests"}, fields[8], mm.GAUGE}
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{device, "io", "ticks"}, fields[9], mm.COUNTER}
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{device, "queue", "time"}, fields[10], mm.COUNTER}
	}
}

func (poller Disk) Name() string {
	return "disk"
}

func feedDevices(devices chan<- string) {
	defer close(devices)
	file, err := os.Open("/proc/partitions")
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

		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] == "major" {
			continue
		}

		if utils.Exists(SYS + fields[3]) {
			devices <- fields[3]
		} else {
			continue
		}
	}
}
