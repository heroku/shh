package pollers

import (
	"github.com/freeformz/shh/mm"
	"github.com/freeformz/shh/utils"
	"sort"
	"strings"
	"time"
)

const (
	DEVICE_FILE     = "/proc/net/dev"
	DEFAULT_DEVICES = "eth0,lo"
)

var (
	devices = utils.GetEnvWithDefaultStrings("SHH_NIF_DEVICES", DEFAULT_DEVICES)
)

type NetworkInterface struct {
	measurements chan<- *mm.Measurement
}

func NewNetworkInterfacePoller(measurements chan<- *mm.Measurement) NetworkInterface {
	return NetworkInterface{measurements: measurements}
}

// http://www.kernel.org/doc/Documentation/filesystems/proc.txt (section 1.4)
func (poller NetworkInterface) Poll(tick time.Time) {

	for line := range utils.FileLineChannel(DEVICE_FILE) {
		fields := strings.Fields(line)
		if strings.HasSuffix(fields[0], ":") {
			device := fields[0]
			nakedDevice := device[:len(device)-1]
			idx := sort.SearchStrings(devices, nakedDevice)
			if idx < len(devices) && fields[idx] == device {
				// It's a device we want to gather metrics for

				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "receive", "bytes"}, utils.Atouint64(fields[1])}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "receive", "packets"}, utils.Atouint64(fields[2])}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "receive", "errors"}, utils.Atouint64(fields[3])}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "receive", "dropped"}, utils.Atouint64(fields[4])}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "receive", "errors", "fifo"}, utils.Atouint64(fields[5])}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "receive", "errors", "frame"}, utils.Atouint64(fields[6])}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "receive", "compressed"}, utils.Atouint64(fields[7])}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "receive", "multicast"}, utils.Atouint64(fields[8])}

				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "transmit", "bytes"}, utils.Atouint64(fields[9])}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "transmit", "packets"}, utils.Atouint64(fields[10])}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "transmit", "errors"}, utils.Atouint64(fields[11])}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "transmit", "dropped"}, utils.Atouint64(fields[12])}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "transmit", "errors", "fifo"}, utils.Atouint64(fields[13])}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "transmit", "errors", "collisions"}, utils.Atouint64(fields[14])}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "transmit", "errors", "carrier"}, utils.Atouint64(fields[15])}
				poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "transmit", "compressed"}, utils.Atouint64(fields[16])}

			}
		}
	}
}

func (poller NetworkInterface) Name() string {
	return "nif"
}
