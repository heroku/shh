package shh

import (
	"fmt"
)

type StdOutL2MetRaw struct {
	measurements <-chan *Measurement
}

func NewStdOutL2MetRaw(measurements <-chan *Measurement) StdOutL2MetRaw {
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
	incoming <-chan *Measurement
	outgoing chan<- *Measurement
	last     map[string]*Measurement
	raw      StdOutL2MetRaw
}

func NewStdOutL2MetDer(measurements <-chan *Measurement) StdOutL2MetDer {
	plex := make(chan *Measurement)
	return StdOutL2MetDer{measurements, plex, make(map[string]*Measurement), StdOutL2MetRaw{plex}}
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
					out.outgoing <- &Measurement{measurement.When, measurement.Poller, measurement.What, measurement.Difference(last)}
				}
				out.last[key] = measurement
			}
		default:
			out.outgoing <- measurement
		}
	}
}
