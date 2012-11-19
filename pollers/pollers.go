package pollers

import (
	"fmt"
	"shh/mm"
	"shh/pollers/load"
  "shh/pollers/memory"
	"time"
)

type PollerFunc func(now time.Time, measurements chan *mm.Measurement)

var pollers = make(map[string]PollerFunc)

func RegisterPoller(name string, f PollerFunc) {
	pollers[name] = f
}

func Poll(now time.Time, measurements chan *mm.Measurement) {
	for name, pollerFunc := range pollers {
		measurements <- &mm.Measurement{now, fmt.Sprintf("ticking.%s", name), []byte("true")}
		go pollerFunc(now, measurements)
	}
}

func init() {
	RegisterPoller(load.Name, load.Poll)
  RegisterPoller(memory.Name, memory.Poll)
}
