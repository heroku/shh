package outputters

import (
	"fmt"
	"github.com/freeformz/shh/mm"
)

type Outputter interface {
	Setup()
	Poll(measurements <-chan *mm.Measurement)
}

type L2MetStdOut struct {
	Outputter
}

func (out L2MetStdOut) Start(measurements <-chan *mm.Measurement) {
	go out.Output(measurements)
}

func (out L2MetStdOut) Output(measurements <-chan *mm.Measurement) {
	for measurement := range measurements {
		fmt.Println(measurement)
	}
}
