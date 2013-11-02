package shh

import (
	"fmt"
	"net"
	"strconv"
)

type Statsd struct {
	measurements <-chan *Measurement
	last         map[string]*Measurement
	Proto        string
	Host         string
}

func NewStatsdOutputter(measurements <-chan *Measurement, config Config) Statsd {
	return Statsd{
		measurements: measurements,
		last:         make(map[string]*Measurement),
		Proto:        config.StatsdProto,
		Host:         config.StatsdHost,
	}
}

func (out Statsd) Start() {
	go out.Output()
}

func (out Statsd) Connect(host string) net.Conn {
	ctx := Slog{"fn": "Connect", "outputter": "statsd"}

	conn, err := net.Dial(out.Proto, host)
	if err != nil {
		ctx.FatalError(err, "Connecting to statsd host")
	}

	return conn
}

func (s Statsd) Encode(measurement *Measurement) string {
	switch measurement.Value.(type) {
	case uint64:
		key := measurement.Measured()
		last, ok := s.last[key]
		s.last[key] = measurement
		if ok {
			return fmt.Sprintf("%s:%s|c", key, strconv.FormatUint(measurement.Difference(last), 10))
		}
	case float64:
		return fmt.Sprintf("%s:%s|g", measurement.Measured(), measurement.SValue())
	}
	return ""
}

func (out Statsd) Output() {

	conn := out.Connect(out.Host)

	for measurement := range out.measurements {
		fmt.Fprintf(conn, out.Encode(measurement))
	}
}
