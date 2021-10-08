package shh

import (
	"io/ioutil"
	"strings"
	"time"

	"github.com/heroku/shh/filechan"
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

// handlePercentCpu parses a line from cpuacct.stat like "user 12345678",
// calculates the delta from the last measurement, calculates the
// average percentage of one CPU core used, and submits the data point.
func (poller Cgroup) handlePercentCpu(line string, tick time.Time, cgroup string) {
	fields := strings.Fields(line)

	// "user" or "system"
	metric := fields[0]

	// absolute number of centiseconds of CPU time used
	centis := Atouint64(fields[1])

	key := cgroup + "." + metric
	last, exists := poller.last[key]

	if exists {
		delta := centis - last
		percent := float64(delta) * 100.0 / float64(poller.totalCentis)

		poller.measurements <- FloatGaugeMeasurement{tick, poller.Name(), []string{sanitizeMetricName(cgroup), "cpu", metric}, percent, Percent}
	}

	poller.last[key] = centis
}

// handleMaxMemory reads one kind of memory high-water mark, emits a metric,
// and resets the HWM for the next interval.
func (poller Cgroup) handleMaxMemory(metric string, fileName string, tick time.Time, cgroup string) {
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
	for _, cgroup := range poller.cgroups {
		// I can't use the FileLineChannel in utils.go here because I
		// don't want to raise a fatal error if the cgroup doesn't exist yet.

		cpuStat, err := filechan.FileLineChannel(CGROUPS_PATH + "/cpuacct/" + cgroup + "/cpuacct.stat")

		if err == nil {
			for line := range cpuStat {
				poller.handlePercentCpu(line, tick, cgroup)
			}
		}

		poller.handleMaxMemory("user", "memory.max_usage_in_bytes", tick, cgroup)
		poller.handleMaxMemory("kernel", "memory.kmem.max_usage_in_bytes", tick, cgroup)
		poller.handleMaxMemory("kernel.tcp", "memory.kmem.tcp.max_usage_in_bytes", tick, cgroup)
	}
}

func (poller Cgroup) Name() string {
	return "cgroup"
}

func (poller Cgroup) Exit() {}
