package pollers

import (
	"github.com/freeformz/shh/mm"
	"strconv"
	"time"
)

type Poller interface {
	Name() string
	Poll(tick time.Time)
}

func NewMultiPoller(measurements chan<- *mm.Measurement) Multi {
	return Multi{pollers: make(map[string]Poller), measurements: measurements, counts: make(map[string]int)}
}

type Multi struct {
	measurements chan<- *mm.Measurement
	pollers      map[string]Poller
	counts       map[string]int
}

func (p Multi) RegisterPoller(poller Poller) {
	p.pollers[poller.Name()] = poller
	p.counts[poller.Name()] = 0
}

func (p Multi) Poll(tick time.Time) {
	for name, poller := range p.pollers {
		p.counts[name] += 1
		p.measurements <- &mm.Measurement{tick, p.Name(), []string{"ticks", name, "count"}, strconv.Itoa(p.counts[name]), mm.COUNTER}
		go poller.Poll(tick)
	}
}

func (p Multi) Name() string {
	return "multi_poller"
}
