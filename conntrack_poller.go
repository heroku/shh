package main

import (
	"bytes"
	"io/ioutil"
	"time"
)

const (
	CONNTRACK_DATA = "/proc/sys/net/netfilter/nf_conntrack_count"
)

type Conntrack struct {
	measurements chan<- Measurement
}

func NewConntrackPoller(measurements chan<- Measurement) Conntrack {
	return Conntrack{measurements: measurements}
}

func (poller Conntrack) Poll(tick time.Time) {
	ctx := Slog{"poller": poller.Name(), "fn": "Poll", "tick": tick}

	count, err := ioutil.ReadFile(CONNTRACK_DATA)
	if err != nil {
		ctx.Error(err, "reading "+CONNTRACK_DATA)
	}

	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"count"}, Atouint64(string(bytes.TrimSpace(count)))}
}

func (poller Conntrack) Name() string {
	return "conntrack"
}

func (poller Conntrack) Exit() {}
