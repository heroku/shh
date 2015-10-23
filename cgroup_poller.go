package shh

import (
	"io/ioutil"
	"strings"
	"time"

	"github.com/heroku/shh/Godeps/_workspace/src/github.com/freeformz/filechan"
)

const (
	CGROUPS_PATH = "/sys/fs/cgroup"
)

type Cgroup struct {
	measurements chan<- Measurement
	last         map[string]uint64
	cgroups      []string
	// The kernel will report CPU usage in centiseconds.  This
	// stores the total centiseconds in the polling interval.
	totalCentis uint64
}

func NewCgroupPoller(measurements chan<- Measurement, config Config) Cgroup {
	println("constructed")
	return Cgroup{
		measurements: measurements,
		last:         make(map[string]uint64),
		cgroups:      config.Cgroups,
		// convert the interval to centiseconds
		totalCentis: uint64(config.Interval.Nanoseconds() / 10000000),
	}
}

func sanitizeMetricName(name string) string {
	return strings.Replace(strings.Replace(name, ".", "-", -1), "/", "--", -1)
}

// parsePercent parses a line from cpuacct.stat like "user 12345678",
// calculates the delta from the last measurement, calculates the
// average percentage of one CPU core used, and submits the data point.
func (poller Cgroup) parsePercentCpu(line string, tick time.Time, cgroup string) {
	println("cpu " + line)

	fields := strings.Fields(line)

	// "user" or "system"
	metric := fields[0]

	// absolute number of centiseconds of CPU time used
	centis := Atouint64(fields[1])

	last, exists := poller.last[metric]

	if exists {
		delta := centis - last
		percent := float64(delta) * 100.0 / float64(poller.totalCentis)

		poller.measurements <- FloatGaugeMeasurement{tick, poller.Name(), []string{sanitizeMetricName(cgroup), "cpu", metric}, percent, Percent}
	}

	poller.last[metric] = centis
}

func (poller Cgroup) parseMaxMemory(metric string, fileName string, tick time.Time, cgroup string) {
	println("memory " + metric)

	path := CGROUPS_PATH + "/memory/" + cgroup + "/" + fileName
	data, err := ioutil.ReadFile(path)

	if err == nil {
		maxBytes := Atofloat64(strings.TrimSpace(string(data)))
		poller.measurements <- FloatGaugeMeasurement{tick, poller.Name(), []string{sanitizeMetricName(cgroup), "mem", metric}, maxBytes, Bytes}

		// reset the high water mark
		ioutil.WriteFile(path, []byte("0"), 0644)
	}
}

func (poller Cgroup) Poll(tick time.Time) {
	println("poll")

	for _, cgroup := range poller.cgroups {
		println(cgroup)

		// I can't use the FileLineChannel here because I don't want
		// to raise a fatal error if the cgroup doesn't exist yet.

		cpuStat, err := filechan.FileLineChannel(CGROUPS_PATH + "/cpuacct/" + cgroup + "/cpuacct.stat")

		if err == nil {
			for line := range cpuStat {
				poller.parsePercentCpu(line, tick, cgroup)
			}
		}

		poller.parseMaxMemory("user", "memory.max_usage_in_bytes", tick, cgroup)
		poller.parseMaxMemory("kernel", "memory.kmem.max_usage_in_bytes", tick, cgroup)
		poller.parseMaxMemory("kernel.tcp", "memory.kmem.tcp.max_usage_in_bytes", tick, cgroup)
	}
}

func (poller Cgroup) Name() string {
	return "cgroup"
}

func (poller Cgroup) Exit() {}
