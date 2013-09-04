package output

import (
	"fmt"
	"github.com/freeformz/shh/config"
	"github.com/freeformz/shh/mm"
	"github.com/freeformz/shh/utils"
	"net"
	"strconv"
)

type Statsd struct {
	measurements <-chan *mm.Measurement
	last map[string]*mm.Measurement
}

func NewStatsdOutputter(measurements <-chan *mm.Measurement) Statsd {
	return Statsd{measurements, make(map[string]*mm.Measurement)}
}

func (out Statsd) Start() {
	go out.Output()
}

func (out Statsd) Connect(host string) net.Conn {
	ctx := utils.Slog{"fn": "Connect", "outputter": "statsd"}

	conn, err := net.Dial(config.StatsdProto, host)
	if err != nil {
		ctx.FatalError(err, "Connecting to statsd host")
	}

	return conn
}

func (s Statsd) Encode(measurement *mm.Measurement) string {
	switch measurement.Value.(type) {
	case uint64:
		key := measurement.Measured()
		last, ok := s.last[key]
		s.last[key] = measurement
		if ok {
			return fmt.Sprintf("%s:%s|c", key, strconv.FormatUint(measurement.Difference(last),10))
		} 
	case float64:
		return fmt.Sprintf("%s:%s|g", measurement.Measured(), measurement.SValue())
	}
	return ""
}

func (out Statsd) Output() {

	conn := out.Connect(config.StatsdHost)

	for measurement := range out.measurements {
		fmt.Fprintf(conn, out.Encode(measurement))
	}
}
