package main

import (
	"bufio"
	"os"
	"strings"
	"time"
)

const (
	LOAD_DATA = "/proc/loadavg"
)

type Load struct {
	measurements chan<- *Measurement
}

func NewLoadPoller(measurements chan<- *Measurement) Load {
	return Load{measurements: measurements}
}

func (poller Load) Poll(tick time.Time) {
	ctx := Slog{"poller": poller.Name(), "fn": "Poll", "tick": tick}

	file, err := os.Open(LOAD_DATA)
	if err != nil {
		ctx.FatalError(err, "opening "+LOAD_DATA)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	line, err := reader.ReadString('\n')
	if err != nil {
		ctx.FatalError(err, "reading line from "+LOAD_DATA)
	}
	fields := strings.Fields(line)
	poller.measurements <- &Measurement{tick, poller.Name(), []string{"1m"}, Atofloat64(fields[0])}
	poller.measurements <- &Measurement{tick, poller.Name(), []string{"5m"}, Atofloat64(fields[1])}
	poller.measurements <- &Measurement{tick, poller.Name(), []string{"15m"}, Atofloat64(fields[2])}
	entities := strings.Split(fields[3], "/")
	poller.measurements <- &Measurement{tick, poller.Name(), []string{"scheduling", "entities", "executing"}, Atofloat64(entities[0])}
	poller.measurements <- &Measurement{tick, poller.Name(), []string{"scheduling", "entities", "total"}, Atofloat64(entities[1])}
	poller.measurements <- &Measurement{tick, poller.Name(), []string{"pid", "last"}, Atofloat64(fields[4])}
}

func (poller Load) Name() string {
	return "load"
}

func (poller Load) Exit() {}
