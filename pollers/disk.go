package pollers

import (
	"bufio"
	"github.com/freeformz/shh/mm"
	"github.com/freeformz/shh/utils"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	SYS = "/sys/block/"
)

type Disk struct{}

// http://www.kernel.org/doc/Documentation/block/stat.txt
func (poller Disk) Poll(measurements chan<- *mm.Measurement) {
	devices := make(chan string)
	go feedDevices(devices)

	for device := range devices {
		statBytes, err := ioutil.ReadFile(SYS + device + "/stat")
		if err != nil {
			log.Fatal(err)
		}

		fields := strings.Fields(string(statBytes))
		measurements <- &mm.Measurement{poller.Name(), []string{device, "read", "requests"}, fields[0]}
		measurements <- &mm.Measurement{poller.Name(), []string{device, "read", "merges"}, fields[1]}
		measurements <- &mm.Measurement{poller.Name(), []string{device, "read", "sectors"}, fields[2]}
		measurements <- &mm.Measurement{poller.Name(), []string{device, "read", "ticks"}, fields[3]}

		measurements <- &mm.Measurement{poller.Name(), []string{device, "write", "requests"}, fields[4]}
		measurements <- &mm.Measurement{poller.Name(), []string{device, "write", "merges"}, fields[5]}
		measurements <- &mm.Measurement{poller.Name(), []string{device, "write", "sectors"}, fields[6]}
		measurements <- &mm.Measurement{poller.Name(), []string{device, "write", "ticks"}, fields[7]}

		measurements <- &mm.Measurement{poller.Name(), []string{device, "in_flight", "requests"}, fields[8]}
		measurements <- &mm.Measurement{poller.Name(), []string{device, "io", "ticks"}, fields[9]}
		measurements <- &mm.Measurement{poller.Name(), []string{device, "queue", "time"}, fields[10]}
	}
}

func (poller Disk) Name() string {
	return "disk"
}

func feedDevices(devices chan<- string) {
	defer close(devices)
	file, err := os.Open("/proc/partitions")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] == "major" {
			continue
		}

		if utils.Exists(SYS + fields[3]) {
			devices <- fields[3]
		} else {
			continue
		}
	}
}
