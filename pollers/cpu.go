package pollers

import (
	"bufio"
	"bytes"
	"github.com/freeformz/shh/mm"
	"io"
	"log"
	"os"
	"time"
)

type Cpu struct{}

func (poller Cpu) Poll(tick time.Time, measurements chan *mm.Measurement) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		if bytes.HasPrefix(line, []byte("cpu")) {
			fields := bytes.Fields(line)
			cpu := string(fields[0])
			measurements <- &mm.Measurement{tick, cpu + ".user", fields[1]}
			measurements <- &mm.Measurement{tick, cpu + ".nice", fields[2]}
			measurements <- &mm.Measurement{tick, cpu + ".system", fields[3]}
			measurements <- &mm.Measurement{tick, cpu + ".idle", fields[4]}
			measurements <- &mm.Measurement{tick, cpu + ".iowait", fields[5]}
			measurements <- &mm.Measurement{tick, cpu + ".irq", fields[6]}
			measurements <- &mm.Measurement{tick, cpu + ".softirq", fields[7]}
			measurements <- &mm.Measurement{tick, cpu + ".steal", fields[8]}
			measurements <- &mm.Measurement{tick, cpu + ".guest", fields[9]}
		}
	}
}

func (poller Cpu) Name() string {
	return "cpu"
}
