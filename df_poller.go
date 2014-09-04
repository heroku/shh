package main

import (
	"strings"
	"syscall"
	"time"
)

type Df struct {
	measurements chan<- Measurement
	percentage   bool
	Types        []string
}

func NewDfPoller(measurements chan<- Measurement, config Config) Df {
	return Df{
		measurements: measurements,
		percentage:   LinearSliceContainsString(config.Percentages, "df"),
		Types:        config.DfTypes,
	}
}

func (poller Df) Poll(tick time.Time) {
	ctx := Slog{"poller": poller.Name(), "fn": "Poll", "tick": tick}

	buf := new(syscall.Statfs_t)

	for mp := range poller.mountpointChannel() {
		err := syscall.Statfs(mp, buf)
		if err != nil {
			ctx["mountpoint"] = mp
			ctx.Error(err, "calling Statfs")
			poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"error"}, 1}
			continue
		}
		mmp := massageMountPoint(mp)
		total_bytes := uint64(buf.Bsize) * buf.Blocks
		user_free_bytes := uint64(buf.Bsize) * buf.Bavail
		root_free_bytes := uint64(buf.Bsize) * buf.Bfree - user_free_bytes
		used_bytes := total_bytes - root_free_bytes - user_free_bytes

		poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{mmp, "total_bytes"}, total_bytes}
		poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{mmp, "root", "free", "bytes"}, root_free_bytes}
		poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{mmp, "user", "free", "bytes"}, user_free_bytes}
		poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{mmp, "used", "bytes"}, used_bytes}
		poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{mmp, "total", "inodes"}, buf.Files}
		poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{mmp, "free", "inodes"}, buf.Ffree}

		if poller.percentage {
			poller.measurements <- &FloatGaugeMeasurement{tick, poller.Name(), []string{mmp, "used", "perc"}, float64(used_bytes) / float64(total_bytes)}
		}
	}
}

func (poller Df) Name() string {
	return "df"
}
func (poller Df) Exit() {}

// Utility functions
// Massages the mount point so that "/" == "root" and
// other substitutions
func massageMountPoint(path string) string {
	switch path {
	case "/":
		return "root"
	}
	if strings.HasPrefix(path, "/") {
		path = strings.TrimLeft(path, "/")
	}
	path = strings.Replace(path, "/", "_", -1)
	return path
}

// Returns a channel on which you can receive the mountspoints we care about
func (poller Df) mountpointChannel() <-chan string {
	c := make(chan string)

	go func(mountpoints chan<- string) {
		defer close(mountpoints)

		for line := range FileLineChannel("/proc/mounts") {

			fields := strings.Fields(line)
			fsType := fields[2]

			if SliceContainsString(poller.Types, fsType) {
				mountpoints <- fields[1]
			}
		}
	}(c)

	return c
}
