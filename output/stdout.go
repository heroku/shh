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
	last     map[string]*mm.Measurement
}

func NewStdOutL2MetDer(measurements <-chan *mm.Measurement) StdOutL2MetDer {
	return StdOutL2MetDer{measurements, make(map[string]*mm.Measurement)}
}

func (out StdOutL2MetDer) Start() {
	go out.Output()
}

func (out StdOutL2MetDer) Output() {
	for measurement := range out.incoming {
		switch measurement.Value.(type) {
		case float64:
			{
				fmt.Println(measurement)
			}
		case uint64:
			{
				key := measurement.Measured()
				last := out.last[key]
				if last != nil {
					tmp := &mm.Measurement{measurement.When, measurement.Poller, measurement.What, measurement.Difference(last)}
					fmt.Println(tmp)
				}
				out.last[key] = measurement
			}
		}
	}
}
