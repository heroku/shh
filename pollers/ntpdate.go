package pollers

import (
	"bufio"
	"github.com/freeformz/shh/mm"
	"github.com/freeformz/shh/utils"
	"io"
	"log"
	"os/exec"
	"strings"
	"time"
)

var (
	servers = utils.GetEnvWithDefaultStrings("SHH_NTPDATE_SERVERS", "0.pool.ntp.org,1.pool.ntp.org")
)

type Ntpdate struct {
	measurements chan<- *mm.Measurement
}

func NewNtpdatePoller(measurements chan<- *mm.Measurement) Ntpdate {
	return Ntpdate{measurements: measurements}
}

//FIXME: Timeout
func (poller Ntpdate) Poll(tick time.Time) {
	if len(servers) > 0 {
		cmd := exec.Command("ntpdate", "-q", "-u")
		cmd.Args = append(cmd.Args, servers...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}

		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}

		defer func() {
			if err := cmd.Wait(); err != nil {
				log.Fatal(err)
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
					log.Fatal(err)
				}
			}
		}

	}
}

func (poller Ntpdate) Name() string {
	return "ntpdate"
}

func (poller Ntpdate) Exit() {}
