package shh

import (
	"fmt"
	"net"
	"strconv"
)

type Statsd struct {
	measurements <-chan Measurement
	last         map[string]CounterMeasurement
	Proto        string
	Host         string
	prefix       string
	source       string
}

func NewStatsdOutputter(measurements <-chan Measurement, config Config) *Statsd {
	return &Statsd{
		measurements: measurements,
		last:         make(map[string]CounterMeasurement),
		Proto:        config.StatsdProto,
		Host:         config.StatsdHost,
		prefix:       config.Prefix,
		source:       config.Source, // TODO: unused?
	}
}

func (out *Statsd) Start() {
	go out.Output()
}

func (out *Statsd) Connect(host string) net.Conn {
	ctx := Slog{"fn": "Connect", "outputter": "statsd"}

	conn, err := net.Dial(out.Proto, host)
	if err != nil {
		ctx.FatalError(err, "Connecting to statsd host")
	}

	return conn
}

func (s *Statsd) Encode(mm Measurement) string {
	switch mm.Type() {
	case CounterType:
		key := mm.Name(s.prefix)
		last, ok := s.last[key]
		s.last[key] = mm.(CounterMeasurement)
		if ok {
			return fmt.Sprintf("%s:%s|c", key, strconv.FormatUint(
				mm.(CounterMeasurement).Difference(last), 10))
		}
	case FloatGaugeType, GaugeType:
		return fmt.Sprintf("%s:%s|g", mm.Name(s.prefix), mm.StrValue())
	}
	return ""
}

func (out *Statsd) Output() {
	conn := out.Connect(out.Host)

	for mm := range out.measurements {
		if ms := out.Encode(mm); ms != "" {
			fmt.Fprintf(conn, ms)
		}
	}
}
