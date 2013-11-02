package shh

import (
	"io/ioutil"
	"strings"
	"time"
)

const (
	SYS         = "/sys/block/"
	SECTOR_SIZE = 512
)

type Disk struct {
	measurements chan<- *Measurement
}

func NewDiskPoller(measurements chan<- *Measurement) Disk {
	return Disk{measurements: measurements}
}

// http://www.kernel.org/doc/Documentation/block/stat.txt
func (poller Disk) Poll(tick time.Time) {
	ctx := Slog{"poller": poller.Name(), "fn": "Poll", "tick": tick}

	for device := range deviceChannel() {
		target := SYS + device + "/stat"
		statBytes, err := ioutil.ReadFile(target)
		if err != nil {
			ctx.FatalError(err, "reading "+target)
		}

		fields := strings.Fields(string(statBytes))
		poller.measurements <- &Measurement{tick, poller.Name(), []string{device, "read", "requests"}, Atouint64(fields[0])}
		poller.measurements <- &Measurement{tick, poller.Name(), []string{device, "read", "merges"}, Atouint64(fields[1])}
		poller.measurements <- &Measurement{tick, poller.Name(), []string{device, "read", "bytes"}, Atouint64(fields[2]) * SECTOR_SIZE}
		poller.measurements <- &Measurement{tick, poller.Name(), []string{device, "read", "ticks"}, Atouint64(fields[3])}

		poller.measurements <- &Measurement{tick, poller.Name(), []string{device, "write", "requests"}, Atouint64(fields[4])}
		poller.measurements <- &Measurement{tick, poller.Name(), []string{device, "write", "merges"}, Atouint64(fields[5])}
		poller.measurements <- &Measurement{tick, poller.Name(), []string{device, "write", "bytes"}, Atouint64(fields[6]) * SECTOR_SIZE}
		poller.measurements <- &Measurement{tick, poller.Name(), []string{device, "write", "ticks"}, Atouint64(fields[7])}

		poller.measurements <- &Measurement{tick, poller.Name(), []string{device, "in_flight", "requests"}, Atofloat64(fields[8])}
		poller.measurements <- &Measurement{tick, poller.Name(), []string{device, "io", "ticks"}, Atouint64(fields[9])}
		poller.measurements <- &Measurement{tick, poller.Name(), []string{device, "queue", "time"}, Atouint64(fields[10])}
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

		for line := range FileLineChannel("/proc/partitions") {

			fields := strings.Fields(line)
			if len(fields) == 0 || fields[0] == "major" {
				continue
			}

			if Exists(SYS + fields[3]) {
				devices <- fields[3]
			} else {
				continue
			}
		}
	}(c)

	return c
}
