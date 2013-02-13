package main

import (
	"fmt"
	"github.com/freeformz/shh/mm"
	"github.com/freeformz/shh/output"
	"github.com/freeformz/shh/pollers"
	"github.com/freeformz/shh/utils"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	VERSION           = "0.0.23"
	DEFAULT_INTERVAL  = "10s"            // Default tick interval for pollers
	DEFAULT_OUTPUTTER = "stdoutl2metder" // Default outputter
)

var (
	Interval      = utils.GetEnvWithDefaultDuration("SHH_INTERVAL", DEFAULT_INTERVAL) // Polling Interval
	Outputter     = utils.GetEnvWithDefault("SHH_OUTPUTTER", DEFAULT_OUTPUTTER)       // Outputter
	Start         = time.Now()                                                        // Start time
	ProfilePort   = utils.GetEnvWithDefault("SHH_PROFILE_PORT", "0")                  // Profile Port
	SignalChannel = make(chan os.Signal, 1)
)

func main() {
	measurements := make(chan *mm.Measurement, 100)

	mp := pollers.NewMultiPoller(measurements)

	signal.Notify(SignalChannel, syscall.SIGINT)
	signal.Notify(SignalChannel, syscall.SIGTERM)

	go func() {
		for sig := range SignalChannel {
			mp.Exit()
			log.Fatalf("signal=%s finished=%s duration=%s\n", sig, time.Now().Format(time.RFC3339Nano), time.Since(Start))
		}
	}()

	if ProfilePort != "0" {
		go func() {
			log.Println(http.ListenAndServe("localhost:"+ProfilePort, nil))
		}()
	}

	fmt.Printf("shh_start=true at=%s interval=%s\n", Start.Format(time.RFC3339Nano), Interval)

	outputter, err := output.NewOutputter(Outputter, measurements)
	if err != nil {
		log.Fatal(err)
	}
	outputter.Start()

	// poll at start
	go mp.Poll(time.Now())

	ticks := time.Tick(Interval)
	for {
		select {
		case tick := <-ticks:
			go mp.Poll(tick)
		}
	}
}
