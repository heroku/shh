package main

import (
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/heroku/slog"
)

const (
	SYS         = "/sys/block/"
	SECTOR_SIZE = 512
)

type Disk struct {
	measurements chan<- Measurement
	diskFilter   *regexp.Regexp
}

func NewDiskPoller(measurements chan<- Measurement, config Config) Disk {
	return Disk{measurements: measurements, diskFilter: config.DiskFilter}
}

// http://www.kernel.org/doc/Documentation/block/stat.txt
func (poller Disk) Poll(tick time.Time) {
	ctx := slog.Context{"poller": poller.Name(), "fn": "Poll", "tick": tick}

	for device := range deviceChannel(poller.diskFilter) {
		target := SYS + device + "/stat"
		statBytes, err := ioutil.ReadFile(target)
		if err != nil {
			LogError(ctx, err, "reading"+target)
			poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"error"}, 1, Errors}
			continue
		}

		fields := strings.Fields(string(statBytes))
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{device, "read", "requests"}, Atouint64(fields[0]), Requests}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{device, "read", "merges"}, Atouint64(fields[1]), Requests}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{device, "read", "bytes"}, Atouint64(fields[2]) * SECTOR_SIZE, Bytes}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{device, "read", "ticks"}, Atouint64(fields[3]), MilliSeconds}

		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{device, "write", "requests"}, Atouint64(fields[4]), Requests}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{device, "write", "merges"}, Atouint64(fields[5]), Requests}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{device, "write", "bytes"}, Atouint64(fields[6]) * SECTOR_SIZE, Bytes}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{device, "write", "ticks"}, Atouint64(fields[7]), MilliSeconds}

		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{device, "in_flight", "requests"}, Atouint64(fields[8]), Requests}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{device, "io", "ticks"}, Atouint64(fields[9]), MilliSeconds}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{device, "queue", "time"}, Atouint64(fields[10]), MilliSeconds}
	}

}

func (poller Disk) Name() string {
	return "disk"
}
func (poller Disk) Exit() {}

func deviceChannel(filter *regexp.Regexp) <-chan string {
	c := make(chan string)

	go func(devices chan<- string) {
		defer close(devices)

		for line := range FileLineChannel("/proc/partitions") {

			fields := strings.Fields(line)
			if len(fields) == 0 || fields[0] == "major" {
				continue
			}

			if Exists(SYS+fields[3]) && filter.MatchString(fields[3]) {
				devices <- fields[3]
			} else {
				continue
			}
		}
	}(c)

	return c
}
