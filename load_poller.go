package shh

import (
	"bufio"
	"os"
	"strings"
	"time"

	"github.com/heroku/shh/Godeps/_workspace/src/github.com/heroku/slog"
)

const (
	LOAD_DATA = "/proc/loadavg"
)

type Load struct {
	measurements chan<- Measurement
}

func NewLoadPoller(measurements chan<- Measurement) Load {
	return Load{measurements: measurements}
}

func (poller Load) Poll(tick time.Time) {
	ctx := slog.Context{"poller": poller.Name(), "fn": "Poll", "tick": tick}

	file, err := os.Open(LOAD_DATA)
	if err != nil {
		FatalError(ctx, err, "opening "+LOAD_DATA)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	line, err := reader.ReadString('\n')
	if err != nil {
		LogError(ctx, err, "reading line from "+LOAD_DATA)
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"error"}, 1, Errors}
		return
	}
	fields := strings.Fields(line)
	poller.measurements <- FloatGaugeMeasurement{tick, poller.Name(), []string{"1m"}, Atofloat64(fields[0]), Avg}
	poller.measurements <- FloatGaugeMeasurement{tick, poller.Name(), []string{"5m"}, Atofloat64(fields[1]), Avg}
	poller.measurements <- FloatGaugeMeasurement{tick, poller.Name(), []string{"15m"}, Atofloat64(fields[2]), Avg}
	entities := strings.Split(fields[3], "/")
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"scheduling", "entities", "executing"}, Atouint64(entities[0]), Processes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"scheduling", "entities", "total"}, Atouint64(entities[1]), Processes}
}

func (poller Load) Name() string {
	return "load"
}

func (poller Load) Exit() {}
