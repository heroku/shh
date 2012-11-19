package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"time"
  "shh/mm"
)

func writeOut(measurements chan *mm.Measurement) {
	for measurement := range measurements {
		fmt.Println(measurement)
	}
}

func pollLoad(now time.Time, measurements chan *mm.Measurement) {
	data, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		log.Fatal(err)
	}
	fields := bytes.Fields(data)
	measurements <- &mm.Measurement{now, "load.1m", fields[0]}
	measurements <- &mm.Measurement{now, "load.5m", fields[1]}
	measurements <- &mm.Measurement{now, "load.15m", fields[2]}
  entities := bytes.Split(fields[3],[ ]byte("/"))
  measurements <- &mm.Measurement{now, "scheduling.entities.executing", entities[0]}
  measurements <- &mm.Measurement{now, "scheduling.entities.total", entities[1]}
  measurements <- &mm.Measurement{now, "pid.last", fields[4]}
	return
}

func main() {
	measurements := make(chan *mm.Measurement, 100)
	duration, _ := time.ParseDuration("5s")
	ticks := time.Tick(duration)
	go writeOut(measurements)
	for now := range ticks {
		measurements <- &mm.Measurement{now, "tick", []byte("true")}
    go pollLoad(now, measurements)
	}
}
