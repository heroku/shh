package output

import (
	"fmt"
	"github.com/freeformz/shh/config"
	"github.com/freeformz/shh/mm"
	"github.com/freeformz/shh/utils"
	"net"
	"strings"
)

type Carbon struct {
	measurements <-chan *mm.Measurement
}

func NewCarbonOutputter(measurements <-chan *mm.Measurement) Carbon {
	return Carbon{measurements}
}

func (out Carbon) Start() {
	go out.Output()
}

func (out Carbon) Connect(host string) net.Conn {
	ctx := utils.Slog{"fn": "Connect", "outputter": "carbon"}

	conn, err := net.Dial("tcp", host)
	if err != nil {
		ctx.FatalError(err, "Connecting to carbon host")
	}

	return conn
}

func (out Carbon) Output() {

	conn := out.Connect(config.CarbonHost)

	metric := make([]string, 0, 10)
	var resetEnd int

	if config.Prefix != "" {
		resetEnd = 1
		metric = append(metric, config.Prefix)
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
