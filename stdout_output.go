package main

import (
	"fmt"
	"time"
)

type StdOutL2MetRaw struct {
	measurements <-chan Measurement
	prefix       string
	source       string
}

func NewStdOutL2MetRaw(measurements <-chan Measurement, config Config) *StdOutL2MetRaw {
	return &StdOutL2MetRaw{measurements: measurements, prefix: config.Prefix, source: config.Source}
}

func (out *StdOutL2MetRaw) Start() {
	go out.Output()
}

func (out *StdOutL2MetRaw) Output() {
	for mm := range out.measurements {
		msg := fmt.Sprintf("when=%s sample#%s=%s", mm.Time().Format(time.RFC3339), mm.Name(out.prefix), mm.StrValue())
		if out.source != "" {
			fmt.Println(fmt.Sprintf("%s source=%s", msg, out.source))
			continue
		}
		fmt.Println(msg)
	}
}

type StdOutL2MetDer struct {
	incoming <-chan Measurement
	outgoing chan<- Measurement
	last     map[string]*CounterMeasurement
	raw      *StdOutL2MetRaw
	prefix   string
}

func NewStdOutL2MetDer(measurements <-chan Measurement, config Config) *StdOutL2MetDer {
	plex := make(chan Measurement)
	return &StdOutL2MetDer{
		incoming: measurements,
		outgoing: plex,
		last:     make(map[string]*CounterMeasurement),
		raw:      NewStdOutL2MetRaw(plex, config),
		prefix:   config.Prefix,
	}
}

func (out *StdOutL2MetDer) Start() {
	go out.Output()
	go out.raw.Output()
}

func (out *StdOutL2MetDer) Output() {
	for mm := range out.incoming {
		switch mm.Type() {
		case CounterType:
			key := mm.Name(out.prefix)
			last, found := out.last[key]
			cm := mm.(*CounterMeasurement)
			out.last[key] = cm
			if found {
				out.outgoing <- CounterMeasurement{cm.time, cm.poller, cm.what, cm.Difference(last)}
			}
		default:
			out.outgoing <- mm
		}
	}
}
