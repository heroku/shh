package pollers

import (
	"github.com/freeformz/shh/mm"
	"github.com/freeformz/shh/utils"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

const (
	SYS         = "/sys/block/"
	SECTOR_SIZE = 512
)

type Disk struct {
	measurements chan<- *mm.Measurement
}

func NewDiskPoller(measurements chan<- *mm.Measurement) Disk {
	return Disk{measurements: measurements}
}

// http://www.kernel.org/doc/Documentation/block/stat.txt
func (poller Disk) Poll(tick time.Time) {

	for device := range deviceChannel() {
		statBytes, err := ioutil.ReadFile(SYS + device + "/stat")
		if err != nil {
			log.Fatal(err)
		}

		fields := strings.Fields(string(statBytes))
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{device, "read", "requests"}, utils.Atouint64(fields[0])}
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{device, "read", "merges"}, utils.Atouint64(fields[1])}
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{device, "read", "bytes"}, utils.Atouint64(fields[2]) * SECTOR_SIZE}
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{device, "read", "ticks"}, utils.Atouint64(fields[3])}

		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{device, "write", "requests"}, utils.Atouint64(fields[4])}
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{device, "write", "merges"}, utils.Atouint64(fields[5])}
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{device, "write", "bytes"}, utils.Atouint64(fields[6]) * SECTOR_SIZE}
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{device, "write", "ticks"}, utils.Atouint64(fields[7])}

		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{device, "in_flight", "requests"}, utils.Atofloat64(fields[8])}
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{device, "io", "ticks"}, utils.Atouint64(fields[9])}
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{device, "queue", "time"}, utils.Atouint64(fields[10])}
	}

}

func (poller Disk) Name() string {
	return "disk"
}
func (poller Disk) Exit() {}

func deviceChannel() <-chan string {
	c := make(chan string)

	go func(devices chan<- string) {
		defer close(devices)

		for line := range utils.FileLineChannel("/proc/partitions") {

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
	}(c)

	return c
}
