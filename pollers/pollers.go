package pollers

import (
	"github.com/freeformz/shh/mm"
	"strconv"
)

type PollerFunc func(measurements chan<- *mm.Measurement)

type Poller interface {
	Name() string
	Poll(measurements chan<- *mm.Measurement)
}

func NewMultiPoller() Multi {
	return Multi{pollers: make(map[string]Poller), counts: make(map[string]int)}
}

type Multi struct {
	pollers map[string]Poller
	counts  map[string]int
}

func (p Multi) RegisterPoller(poller Poller) {
	p.pollers[poller.Name()] = poller
	p.counts[poller.Name()] = 0
}

func (p Multi) Poll(measurements chan<- *mm.Measurement) {
	for name, poller := range p.pollers {
		p.counts[name] += 1
		measurements <- &mm.Measurement{p.Name(), []string{"ticks", name, "count"}, strconv.Itoa(p.counts[name])}
		go poller.Poll(measurements)
	}
}

func (p Multi) Name() string {
	return "multi_poller"
}
