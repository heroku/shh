package output

import (
	"fmt"
	"github.com/freeformz/shh/mm"
)

type StdOutL2MetRaw struct {
	measurements <-chan *mm.Measurement
}

func NewStdOutL2MetRaw(measurements <-chan *mm.Measurement) StdOutL2MetRaw {
	return StdOutL2MetRaw{measurements}
}

func (out StdOutL2MetRaw) Start() {
	go out.Output()
}

func (out StdOutL2MetRaw) Output() {
	for measurement := range out.measurements {
		fmt.Println(measurement)
	}
}

type StdOutL2MetDer struct {
	incoming <-chan *mm.Measurement
	last     map[string]string
}

func NewStdOutL2MetDer(measurements <-chan *mm.Measurement) StdOutL2MetDer {
	return StdOutL2MetDer{measurements, make(map[string]string)}
}

func (out StdOutL2MetDer) Start() {
	go out.Output()
}

func (out StdOutL2MetDer) Output() {
	for measurement := range out.incoming {
		switch measurement.Type {
		case mm.GAUGE:
			{
				fmt.Println(measurement)
			}
		case mm.COUNTER:
			{
				key := measurement.Measured()
				last := out.last[key]
				if last != "" {
					out := fmt.Sprintf("when=%s measure=%s val=%s", measurement.Timestamp(), key, measurement.Difference(last))
					if measurement.Source() != "" {
						out = fmt.Sprintf("%s source=%s", out, measurement.Source())
					}
					fmt.Println(out)
				}
				out.last[key] = measurement.Value
			}
		}
	}
}
