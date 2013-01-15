package pollers

import (
	"fmt"
	"github.com/freeformz/shh/mm"
	"time"
)

type PollerFunc func(tick time.Time, measurements chan *mm.Measurement)

type Poller interface {
	Name() string
	Poll(tick time.Time, measurements chan *mm.Measurement)
}

func NewMultiPoller() Multi {
	return Multi{pollers: make(map[string]Poller)}
}

type Multi struct {
	pollers map[string]Poller
}

func (p Multi) RegisterPoller(poller Poller) {
	p.pollers[poller.Name()] = poller
}

func (p Multi) Poll(tick time.Time, measurements chan *mm.Measurement) {
	for name, poller := range p.pollers {
		measurements <- &mm.Measurement{tick, fmt.Sprintf("ticking.%s", name), []byte("true")}
		go poller.Poll(tick, measurements)
	}
}
