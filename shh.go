package main

import (
	"fmt"
	"shh/mm"
	"shh/pollers"
	"time"
)

func writeOut(measurements chan *mm.Measurement) {
	for measurement := range measurements {
		fmt.Println(measurement)
	}
}

func main() {
	measurements := make(chan *mm.Measurement, 100)
	duration, _ := time.ParseDuration("5s")
	ticks := time.Tick(duration)
	go writeOut(measurements)
	for now := range ticks {
		measurements <- &mm.Measurement{now, "tick", []byte("true")}
		pollers.Poll(now, measurements)
	}
}