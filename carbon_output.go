package shh

import (
	"fmt"
	"net"

	"github.com/heroku/shh/Godeps/_workspace/src/github.com/heroku/slog"
)

type Carbon struct {
	measurements <-chan Measurement
	Host         string
	prefix       string
	source       string
}

func NewCarbonOutputter(measurements <-chan Measurement, config Config) *Carbon {
	return &Carbon{measurements: measurements, Host: config.CarbonHost, prefix: config.Prefix, source: config.Source}
}

func (out *Carbon) Start() {
	go out.Output()
}

func (out *Carbon) Connect(host string) net.Conn {
	ctx := slog.Context{"fn": "Connect", "outputter": "carbon"}

	conn, err := net.Dial("tcp", host)
	if err != nil {
		FatalError(ctx, err, "Connecting to carbon host")
	}

	return conn
}

func (out *Carbon) Output() {
	var prefix string

	conn := out.Connect(out.Host)

	if out.prefix != "" && out.source != "" {
		prefix = fmt.Sprintf("%s.%s", out.prefix, out.source)
	} else if out.prefix == "" {
		prefix = out.source
	} else if out.source == "" {
		prefix = out.prefix
	}

	for mm := range out.measurements {
		fmt.Fprintf(conn, "%s %s %d\n", mm.Name(prefix), mm.StrValue(), mm.Time().Unix())
	}

}
