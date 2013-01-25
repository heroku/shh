package output

import (
	"errors"
	"github.com/freeformz/shh/mm"
)

type Outputter interface {
	Start()
}

//
// FIXME: Any way to do this with reflect and a map?
func NewOutputter(name string, measurements <-chan *mm.Measurement) (Outputter, error) {
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
			return NewLibratoOutputter(measurements), nil
		}
	}

	return nil, errors.New("unknown outputter")
}
