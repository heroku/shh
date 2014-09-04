package main

import (
	"bufio"
	"io"
	"os/exec"
	"strings"
	"time"
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
	ctx := Slog{"poller": poller.Name(), "fn": "Poll", "tick": tick}

	if len(poller.Servers) > 0 {
		cmd := exec.Command("ntpdate", "-q", "-u")
		cmd.Args = append(cmd.Args, poller.Servers...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			ctx.Error(err, "creating stdout pipe")
			poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"error"}, 1}
			return
		}

		if err := cmd.Start(); err != nil {
			ctx.Error(err, "starting sub command")
			poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"error"}, 1}
			return
		}

		defer func() {
			if err := cmd.Wait(); err != nil {
				ctx.Error(err, "waiting for subcommand to end")
				poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"error"}, 1}
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
					poller.measurements <- &FloatGaugeMeasurement{tick, poller.Name(), []string{"offset", server}, Atofloat64(offset)}
					poller.measurements <- &FloatGaugeMeasurement{tick, poller.Name(), []string{"delay", server}, Atofloat64(delay)}
				}
			} else {
				if err == io.EOF {
					break
				} else {
					ctx.Error(err, "unknown error reading data from subcommand")
					poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"error"}, 1}
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
