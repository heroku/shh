package main

import (
	"strings"
	"time"
)

const (
	FILE_NR_DATA = "/proc/sys/fs/file-nr"
)

// /proc/sys/fs/file-nr reports how many file structs the kernel
// currently has allocated, and the maximum number it will allocate.
// Given that an open file uses a file struct, this gives us insight
// into how many more open files we could have.
//
// Note: file descriptors are a per process concept, and are only
// mildly-related to this poller.

type FileNr struct {
	measurements chan<- Measurement
}

func NewFileNrPoller(measurements chan<- Measurement) FileNr {
	return FileNr{
		measurements: measurements,
	}
}

func (poller FileNr) Poll(tick time.Time) {
	for line := range FileLineChannel(FILE_NR_DATA) {
		fields := strings.Split(strings.Trim(line, "\n"), "\t")
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"alloc"}, Atouint64(fields[0]), Files}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"free"}, Atouint64(fields[1]), Files}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"max"}, Atouint64(fields[2]), Files}
	}
}

func (poller FileNr) Name() string {
	return "filenr"
}

func (poller FileNr) Exit() {}
