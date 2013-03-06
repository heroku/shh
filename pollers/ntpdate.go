package pollers

import (
	"bufio"
	"github.com/freeformz/shh/config"
	"github.com/freeformz/shh/mm"
	"github.com/freeformz/shh/utils"
	"io"
	"os/exec"
	"strings"
	"time"
)

type Ntpdate struct {
	measurements chan<- *mm.Measurement
}

func NewNtpdatePoller(measurements chan<- *mm.Measurement) Ntpdate {
	return Ntpdate{measurements: measurements}
}

//FIXME: Timeout
func (poller Ntpdate) Poll(tick time.Time) {
	ctx := utils.Slog{"poller": poller.Name(), "fn": "Poll", "tick": tick}

	if len(config.NtpdateServers) > 0 {
		cmd := exec.Command("ntpdate", "-q", "-u")
		cmd.Args = append(cmd.Args, config.NtpdateServers...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			ctx.FatalError(err, "creating stdout pipe")
		}

		if err := cmd.Start(); err != nil {
			ctx.FatalError(err, "starting sub command")
		}

		defer func() {
			if err := cmd.Wait(); err != nil {
				ctx.FatalError(err, "waiting for subcommand to end")
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
					poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"offset", server}, utils.Atofloat64(offset)}
					poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"delay", server}, utils.Atofloat64(delay)}
				}
			} else {
				if err == io.EOF {
					break
				} else {
					ctx.FatalError(err, "unknown error reading data from subcommand")
				}
			}
		}

	}
}

func (poller Ntpdate) Name() string {
	return "ntpdate"
}

func (poller Ntpdate) Exit() {}
