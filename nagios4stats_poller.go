package shh

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/heroku/slog"
)

type Nagios4StatsPoller struct {
	measurements chan<- Measurement
	metricNames  []string
}

func NewNagios4StatsPoller(measurements chan<- Measurement, config Config) Nagios4StatsPoller {
	return Nagios4StatsPoller{
		measurements: measurements,
		metricNames:  config.Nagios4MetricNames,
	}
}

func (poller Nagios4StatsPoller) Poll(tick time.Time) {
	ctx := slog.Context{"poller": poller.Name(), "fn": "Poll", "tick": tick}

	if len(poller.metricNames) > 0 {
		cmd := exec.Command("nagios4stats", "-m", "-d", strings.Join(poller.metricNames, ","))
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			LogError(ctx, err, "running sub command: "+string(stderr.Bytes()))
			poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"error"}, 1, Errors}
			return
		}

		data := strings.Split(stdout.String(), "\n")
		if (len(data) - 1) != len(poller.metricNames) {
			LogError(ctx, fmt.Errorf("Length of requested metrics and returned metrics differs: %d vs %d", len(poller.metricNames), len(data)-1), "checking stdout")
		} else {
			for i, name := range poller.metricNames {
				poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{strings.ToLower(name)}, Atouint64(data[i]), Empty}
			}
		}
	}

}

func (poller Nagios4StatsPoller) Name() string {
	return "nagios4stats"
}

func (poller Nagios4StatsPoller) Exit() {}
