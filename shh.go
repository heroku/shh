package main

import (
	"fmt"
	"github.com/freeformz/shh/config"
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
			log.Fatal(utils.Slog{"signal": sig, "finished": time.Now(), "duration": time.Since(config.Start)})
		}
	}()

	if config.ProfilePort != config.DEFAULT_PROFILE_PORT {
		go func() {
			log.Println(http.ListenAndServe("localhost:"+config.ProfilePort, nil))
		}()
	}

	ctx := utils.Slog{"shh_start": true, "at": config.Start.Format(time.RFC3339Nano), "interval": config.Interval}
	fmt.Println(ctx)

	outputter, err := output.NewOutputter(config.Outputter, measurements)
	if err != nil {
		ctx.FatalError(err, "creating outputter")
	}
	outputter.Start()

	ticks := utils.PreTick(config.Interval)
	for tick := range ticks {
		mp.Poll(tick)
	}
}
