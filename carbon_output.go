package shh

import (
	"fmt"
	"net"
	"strings"
)

type Carbon struct {
	measurements <-chan *Measurement
	Host         string
	Prefix       string
}

func NewCarbonOutputter(measurements <-chan *Measurement, config Config) Carbon {
	return Carbon{measurements: measurements, Host: config.CarbonHost, Prefix: config.Prefix}
}

func (out Carbon) Start() {
	go out.Output()
}

func (out Carbon) Connect(host string) net.Conn {
	ctx := Slog{"fn": "Connect", "outputter": "carbon"}

	conn, err := net.Dial("tcp", host)
	if err != nil {
		ctx.FatalError(err, "Connecting to carbon host")
	}

	return conn
}

func (out Carbon) Output() {

	conn := out.Connect(out.Host)

	metric := make([]string, 0, 10)
	var resetEnd int

	if out.Prefix != "" {
		resetEnd = 1
		metric = append(metric, out.Prefix)
	} else {
		resetEnd = 0
	}

	for measurement := range out.measurements {
		if source := measurement.Source(); source != "" {
			metric = append(metric, source)
		}
		metric = append(metric, measurement.Poller)
		metric = append(metric, measurement.What...)
		fmt.Fprintf(conn, "%s %s %d\n", strings.Join(metric, "."), measurement.SValue(), measurement.Unix())
		metric = metric[0:resetEnd]
	}
}
