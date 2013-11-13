package main

import (
	"fmt"
)

type StdOutL2MetRaw struct {
	measurements <-chan *Measurement
	prefix       string
	source       string
}

func NewStdOutL2MetRaw(measurements <-chan *Measurement, config Config) *StdOutL2MetRaw {
	return &StdOutL2MetRaw{measurements: measurements, prefix: config.Prefix, source: config.Source}
}

func (out *StdOutL2MetRaw) Start() {
	go out.Output()
}

func (out *StdOutL2MetRaw) Output() {
	for measurement := range out.measurements {
		msg := fmt.Sprintf("when=%s sample#%s=%s", measurement.Timestamp(), measurement.Measured(out.prefix), measurement.SValue())
		if out.source != "" {
			fmt.Println(fmt.Sprintf("%s source=%s", msg, out.source))
			continue
		}
		fmt.Println(msg)
	}
}

type StdOutL2MetDer struct {
	incoming <-chan *Measurement
	outgoing chan<- *Measurement
	last     map[string]*Measurement
	raw      *StdOutL2MetRaw
	prefix   string
}

func NewStdOutL2MetDer(measurements <-chan *Measurement, config Config) *StdOutL2MetDer {
	plex := make(chan *Measurement)
	return &StdOutL2MetDer{
		incoming: measurements,
		outgoing: plex,
		last:     make(map[string]*Measurement),
		raw:      NewStdOutL2MetRaw(plex, config),
		prefix:   config.Prefix,
	}
}

func (out *StdOutL2MetDer) Start() {
	go out.Output()
	go out.raw.Output()
}

func (out *StdOutL2MetDer) Output() {
	for measurement := range out.incoming {
		switch measurement.Value.(type) {
		case uint64:
			{
				key := measurement.Measured(out.prefix)
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
