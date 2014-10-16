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

	"github.com/heroku/shh"
	"github.com/heroku/slog"
)

var (
	signalChannel = make(chan os.Signal, 1)
	versionFlag   = flag.Bool("version", false, "Display version info and exit")
)

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Println(shh.VERSION)
		os.Exit(0)
	}

	measurements := make(chan shh.Measurement, 100)
	config := shh.GetConfig()

	mp := shh.NewMultiPoller(measurements, config)

	signal.Notify(signalChannel, syscall.SIGINT)
	signal.Notify(signalChannel, syscall.SIGTERM)

	go func() {
		for sig := range signalChannel {
			mp.Exit()
			shh.ErrLogger.Fatal(slog.Context{"signal": sig, "finished": time.Now(), "duration": time.Since(config.Start)})
		}
	}()

	if config.ProfilePort != shh.DEFAULT_PROFILE_PORT {
		go func() {
			shh.Logger.Println(http.ListenAndServe("localhost:"+config.ProfilePort, nil))
		}()
	}

	ctx := slog.Context{"start": true, "interval": config.Interval}
	shh.Logger.Println(ctx)

	outputter, err := shh.NewOutputter(config.Outputter, measurements, config)
	if err != nil {
		shh.FatalError(ctx, err, "creating outputter")
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
