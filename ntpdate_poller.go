package main

import (
	"bufio"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/heroku/slog"
)

type Ntpdate struct {
	measurements chan<- Measurement
	Servers      []string
}

func NewNtpdatePoller(measurements chan<- Measurement, config Config) Ntpdate {
	return Ntpdate{
		measurements: measurements,
		Servers:      config.NtpdateServers,
	}
}

//FIXME: Timeout
func (poller Ntpdate) Poll(tick time.Time) {
	ctx := slog.Context{"poller": poller.Name(), "fn": "Poll", "tick": tick}

	if len(poller.Servers) > 0 {
		cmd := exec.Command("ntpdate", "-q", "-u")
		cmd.Args = append(cmd.Args, poller.Servers...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			LogError(ctx, err, "creating stdout pipe")
			poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"error"}, 1, Errors}
			return
		}

		if err := cmd.Start(); err != nil {
			LogError(ctx, err, "starting sub command")
			poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"error"}, 1, Errors}
			return
		}

		defer func() {
			if err := cmd.Wait(); err != nil {
				LogError(ctx, err, "waiting for subcommand to end")
				poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"error"}, 1, Errors}
			}
		}()

		buf := bufio.NewReader(stdout)

		for {
			line, err := buf.ReadString('\n')
			if err == nil {
				if strings.HasPrefix(line, "server") {
					parts := strings.Split(line, ",")
					server := strings.Replace(strings.Fields(parts[0])[1], ".", "_", 4)
					offset := strings.Fields(parts[2])[1]
					delay := strings.Fields(parts[3])[1]
					poller.measurements <- FloatGaugeMeasurement{tick, poller.Name(), []string{"offset", server}, Atofloat64(offset), Seconds}
					poller.measurements <- FloatGaugeMeasurement{tick, poller.Name(), []string{"delay", server}, Atofloat64(delay), Seconds}
				}
			} else {
				if err == io.EOF {
					break
				} else {
					LogError(ctx, err, "unknown error reading data from subcommand")
					poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"error"}, 1, Errors}
					return
				}
			}
		}

	}
}

func (poller Ntpdate) Name() string {
	return "ntpdate"
}

func (poller Ntpdate) Exit() {}
