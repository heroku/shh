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
	outgoing chan<- *mm.Measurement
	last     map[string]*mm.Measurement
	raw      StdOutL2MetRaw
}

func NewStdOutL2MetDer(measurements <-chan *mm.Measurement) StdOutL2MetDer {
	plex := make(chan *mm.Measurement)
	return StdOutL2MetDer{measurements, plex, make(map[string]*mm.Measurement), StdOutL2MetRaw{plex}}
}

func (out StdOutL2MetDer) Start() {
	go out.Output()
	go out.raw.Output()
}

func (out StdOutL2MetDer) Output() {
	for measurement := range out.incoming {
		switch measurement.Value.(type) {
		case uint64:
			{
				key := measurement.Measured()
				last, found := out.last[key]
				if found {
					out.outgoing <- &mm.Measurement{measurement.When, measurement.Poller, measurement.What, measurement.Difference(last)}
				}
				out.last[key] = measurement
			}
		default:
			out.outgoing <- measurement
		}
	}
}
