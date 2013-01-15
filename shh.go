package main

import (
	"fmt"
	"github.com/freeformz/shh/mm"
	"github.com/freeformz/shh/pollers"
	"log"
	"os"
	"time"
)

const (
	DefaultInterval = "10s"
)

func writeOut(measurements chan *mm.Measurement) {
	for measurement := range measurements {
		fmt.Println(measurement)
	}
}

func getDuration() time.Duration {
	interval := os.Getenv("SSH_INTERVAL")

	if interval == "" {
		interval = DefaultInterval
	}

	duration, err := time.ParseDuration(interval)

	if err != nil {
		log.Fatal("unable to parse SHH_INTERVAL: " + interval)
	}

	return duration
}

func main() {
	duration := getDuration()
	fmt.Printf("duration=%s", duration)

	measurements := make(chan *mm.Measurement, 100)
	ticks := time.Tick(duration)

	go writeOut(measurements)

	mp := pollers.NewMultiPoller()
	mp.RegisterPoller(pollers.Load{})
	mp.RegisterPoller(pollers.Cpu{})

	for now := range ticks {
		measurements <- &mm.Measurement{now, "tick", []byte("true")}
		go mp.Poll(now, measurements)
	}
}
