package main

import (
	"fmt"
	"github.com/freeformz/shh/mm"
	"github.com/freeformz/shh/pollers"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	DefaultInterval = "10s"
)

var (
	start = time.Now()
)

func writeOut(measurements chan *mm.Measurement) {
	for measurement := range measurements {
		fmt.Println(measurement)
	}
}

func getDuration() time.Duration {
	interval := os.Getenv("SHH_INTERVAL")

	if interval == "" {
		interval = DefaultInterval
	}

	duration, err := time.ParseDuration(interval)

	if err != nil {
		log.Fatal("unable to parse SHH_INTERVAL: " + interval)
	}

	return duration
}

func init() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		for sig := range c {
			fmt.Printf("signal=%s finished=%s duration=%s\n", sig, time.Now().Format(time.RFC3339Nano), time.Since(start))
			os.Exit(1)
		}
	}()
}

func main() {
	duration := getDuration()
	fmt.Printf("shh_start=true at=%s interval=%s\n", start.Format(time.RFC3339Nano), duration)

	measurements := make(chan *mm.Measurement, 100)
	go writeOut(measurements)

	mp := pollers.NewMultiPoller()
	mp.RegisterPoller(pollers.Load{})
	mp.RegisterPoller(pollers.Cpu{})

	// do a tick at start
	go mp.Poll(measurements)

	ticks := time.Tick(duration)
	for {
		select {
		case <-ticks:
			go mp.Poll(measurements)
		}
	}
}
