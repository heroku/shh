package main

import (
	"errors"
)

type Outputter interface {
	Start()
}

//
// FIXME: Any way to do this with reflect and a map?
func NewOutputter(name string, measurements <-chan *Measurement, config Config) (Outputter, error) {
	switch name {
	case "stdoutl2metraw":
		{
			return NewStdOutL2MetRaw(measurements), nil
		}
	case "stdoutl2metder":
		{
			return NewStdOutL2MetDer(measurements), nil
		}
	case "librato":
		{
			return NewLibratoOutputter(measurements, config), nil
		}
	case "carbon":
		{
			return NewCarbonOutputter(measurements, config), nil
		}
	case "statsd":
		{
			return NewStatsdOutputter(measurements, config), nil
		}
	}

	return nil, errors.New("unknown outputter")
}
