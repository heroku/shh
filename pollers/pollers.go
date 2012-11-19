package pollers

import (
	"fmt"
	"shh/mm"
	"shh/pollers/load"
	"time"
)

var pollers = make(map[string]func(no time.Time, measurements chan *mm.Measurement))

func RegisterPoller(name string, f func(now time.Time, measurements chan *mm.Measurement)) {
	pollers[name] = f
}

func Poll(now time.Time, measurements chan *mm.Measurement) {
	for name, pollerFunc := range pollers {
		measurements <- &mm.Measurement{now, fmt.Sprintf("ticking.%s", name), []byte("true")}
		go pollerFunc(now, measurements)
	}
}

func init() {
	RegisterPoller("load", load.Poll)
}
