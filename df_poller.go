package shh

import (
	"strings"
	"syscall"
	"time"

	"github.com/heroku/slog"
)

type Df struct {
	measurements chan<- Measurement
	percentage   bool
	Types        []string
	Loop         bool
}

func NewDfPoller(measurements chan<- Measurement, config Config) Df {
	return Df{
		measurements: measurements,
		percentage:   LinearSliceContainsString(config.Percentages, "df"),
		Types:        config.DfTypes,
		Loop:         config.DfLoop,
	}
}

func (poller Df) Poll(tick time.Time) {
	ctx := slog.Context{"poller": poller.Name(), "fn": "Poll", "tick": tick}

	buf := new(syscall.Statfs_t)

	for mp := range poller.mountpointChannel() {
		err := syscall.Statfs(mp, buf)
		if err != nil {
			ctx["mountpoint"] = mp
			LogError(ctx, err, "calling Statfs")
			poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"error"}, 1, Errors}
			continue
		}
		mmp := massageMountPoint(mp)
		total_bytes := uint64(buf.Bsize) * buf.Blocks
		user_free_bytes := uint64(buf.Bsize) * buf.Bavail
		root_free_bytes := uint64(buf.Bsize)*buf.Bfree - user_free_bytes
		used_bytes := total_bytes - root_free_bytes - user_free_bytes

		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{mmp, "total", "bytes"}, total_bytes, Bytes}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{mmp, "root", "free", "bytes"}, root_free_bytes, Bytes}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{mmp, "user", "free", "bytes"}, user_free_bytes, Bytes}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{mmp, "used", "bytes"}, used_bytes, Bytes}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{mmp, "total", "inodes"}, buf.Files, INodes}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{mmp, "free", "inodes"}, buf.Ffree, INodes}

		if poller.percentage {
			poller.measurements <- FloatGaugeMeasurement{tick, poller.Name(), []string{mmp, "used", "perc"}, 100.0 * float64(used_bytes) / float64(total_bytes), Percent}
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
			device := fields[0]

			if SliceContainsString(poller.Types, fsType) {
				if !poller.Loop && strings.Contains(device, "/loop") {
					continue
				}

				mountpoints <- fields[1]
			}
		}
	}(c)

	return c
}
