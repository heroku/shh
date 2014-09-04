package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

const (
	PROC = "/proc"
)

type Processes struct {
	measurements chan<- Measurement
}

func NewProcessesPoller(measurements chan<- Measurement) Processes {
	return Processes{measurements: measurements}
}

func (poller Processes) Poll(tick time.Time) {
	ctx := Slog{"poller": poller.Name(), "fn": "Poll", "tick": tick}

	dir, err := os.Open(PROC)
	if err != nil {
		ctx.FatalError(err, "opening "+PROC)
	}

	defer dir.Close()

	dirs, err := dir.Readdirnames(0)
	if err != nil {
		ctx.FatalError(err, "reading dir names")
	}

	var running, sleeping, waiting, zombie, stopped, paging uint64

	for _, proc := range dirs {

		pid, err := strconv.Atoi(proc)

		// Skip anything that isn't an int or < 1
		if err != nil || pid < 1 {
			continue
		}

		switch ProcessState(pid) {
		case "R":
			running++
		case "S":
			sleeping++
		case "D":
			waiting++
		case "Z":
			zombie++
		case "T":
			stopped++
		case "W":
			paging++
		}
	}

	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"running", "count"}, running}
	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"sleeping", "count"}, sleeping}
	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"waiting", "count"}, waiting}
	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"zombie", "count"}, zombie}
	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"stopped", "count"}, stopped}
	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"paging", "count"}, paging}

}

func (poller Processes) Name() string {
	return "processes"
}

func (poller Processes) Exit() {}

func ProcessState(pid int) string {

	statFile := fmt.Sprintf("%s/%d/stat", PROC, pid)

	d, err := ioutil.ReadFile(statFile)

	// Skip errors and return an empty string
	if err != nil {
		return ""
	}

	fields := Fields(string(d))

	return fields[2]
}
