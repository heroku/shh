package pollers

import (
	"bytes"
	"github.com/freeformz/shh/mm"
	"io/ioutil"
	"log"
	"time"
)

type Load struct{}

func (poller Load) Poll(now time.Time, measurements chan *mm.Measurement) {
	data, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		log.Fatal(err)
	}
	fields := bytes.Fields(data)
	measurements <- &mm.Measurement{now, "load.1m", fields[0]}
	measurements <- &mm.Measurement{now, "load.5m", fields[1]}
	measurements <- &mm.Measurement{now, "load.15m", fields[2]}
	entities := bytes.Split(fields[3], []byte("/"))
	measurements <- &mm.Measurement{now, "scheduling.entities.executing", entities[0]}
	measurements <- &mm.Measurement{now, "scheduling.entities.total", entities[1]}
	measurements <- &mm.Measurement{now, "pid.last", fields[4]}
}

func (poller Load) Name() string {
	return "load"
}
