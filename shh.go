package main

import (
	"fmt"
	"github.com/freeformz/shh/config"
	"github.com/freeformz/shh/mm"
	"github.com/freeformz/shh/output"
	"github.com/freeformz/shh/pollers"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	signalChannel = make(chan os.Signal, 1)
)

func main() {
	measurements := make(chan *mm.Measurement, 100)

	mp := pollers.NewMultiPoller(measurements)

	signal.Notify(signalChannel, syscall.SIGINT)
	signal.Notify(signalChannel, syscall.SIGTERM)

	go func() {
		for sig := range signalChannel {
			mp.Exit()
			log.Fatalf("signal=%s finished=%s duration=%s\n", sig, time.Now().Format(time.RFC3339Nano), time.Since(config.Start))
		}
	}()

	if config.ProfilePort != config.DEFAULT_PROFILE_PORT {
		go func() {
			log.Println(http.ListenAndServe("localhost:"+config.ProfilePort, nil))
		}()
	}

	fmt.Printf("shh_start=true at=%s interval=%s\n", config.Start.Format(time.RFC3339Nano), config.Interval)

	outputter, err := output.NewOutputter(config.Outputter, measurements)
	if err != nil {
		log.Fatal(err)
	}
	outputter.Start()

	// poll at start
	go mp.Poll(time.Now())

	ticks := time.Tick(config.Interval)
	for {
		select {
		case tick := <-ticks:
			go mp.Poll(tick)
		}
	}
}
