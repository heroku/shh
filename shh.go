package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/heroku/slog"
)

var (
	signalChannel = make(chan os.Signal, 1)
	versionFlag   = flag.Bool("version", false, "Display version info and exit")
)

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	measurements := make(chan Measurement, 100)
	config := GetConfig()

	mp := NewMultiPoller(measurements, config)

	signal.Notify(signalChannel, syscall.SIGINT)
	signal.Notify(signalChannel, syscall.SIGTERM)

	go func() {
		for sig := range signalChannel {
			mp.Exit()
			ErrLogger.Fatal(slog.Context{"signal": sig, "finished": time.Now(), "duration": time.Since(config.Start)})
		}
	}()

	if config.ProfilePort != DEFAULT_PROFILE_PORT {
		go func() {
			Logger.Println(http.ListenAndServe("localhost:"+config.ProfilePort, nil))
		}()
	}

	ctx := slog.Context{"start": true, "interval": config.Interval}
	Logger.Println(ctx)

	outputter, err := NewOutputter(config.Outputter, measurements, config)
	if err != nil {
		FatalError(ctx, err, "creating outputter")
	}
	outputter.Start()

	start := make(chan time.Time, 1)
	start <- time.Now()
	ticks := time.Tick(config.Interval)

	for {
		select {
		case tick := <-start:
			mp.Poll(tick)
			start = nil
		case tick := <-ticks:
			mp.Poll(tick)
		}
	}
}
