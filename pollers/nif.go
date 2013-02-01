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

type NifReceiveValues struct {
	Bytes      float64
	Packets    float64
	Errors     float64
	Dropped    float64
	Fifo       float64
	Frame      float64
	Compressed float64
	Multicast  float64
}

type NifTransmitValues struct {
	Bytes      float64
	Packets    float64
	Errors     float64
	Dropped    float64
	Fifo       float64
	Collisions float64
	Carrier    float64
	Compressed float64
}

type NifValues struct {
	Receive  NifReceiveValues
	Transmit NifTransmitValues
}

func (nv NifValues) CalcDifference(last NifValues) NifValues {
	return NifValues{
		Receive: NifReceiveValues{
			Bytes:      nv.Receive.Bytes - last.Receive.Bytes,
			Packets:    nv.Receive.Packets - last.Receive.Packets,
			Errors:     nv.Receive.Errors - last.Receive.Errors,
			Dropped:    nv.Receive.Dropped - last.Receive.Dropped,
			Fifo:       nv.Receive.Fifo - last.Receive.Fifo,
			Frame:      nv.Receive.Frame - last.Receive.Frame,
			Compressed: nv.Receive.Compressed - last.Receive.Compressed,
			Multicast:  nv.Receive.Multicast - last.Receive.Multicast,
		},
		Transmit: NifTransmitValues{
			Bytes:      nv.Transmit.Bytes - last.Transmit.Bytes,
			Packets:    nv.Transmit.Packets - last.Transmit.Packets,
			Errors:     nv.Transmit.Errors - last.Transmit.Errors,
			Dropped:    nv.Transmit.Dropped - last.Transmit.Dropped,
			Fifo:       nv.Transmit.Fifo - last.Transmit.Fifo,
			Collisions: nv.Transmit.Collisions - last.Transmit.Collisions,
			Carrier:    nv.Transmit.Carrier - last.Transmit.Carrier,
			Compressed: nv.Transmit.Compressed - last.Transmit.Compressed,
		},
	}
}

type NetworkInterface struct {
	measurements chan<- *mm.Measurement
	last         map[string]NifValues
}

func NewNetworkInterfacePoller(measurements chan<- *mm.Measurement) NetworkInterface {
	return NetworkInterface{measurements: measurements, last: make(map[string]NifValues)}
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

				current := NifValues{
					Receive: NifReceiveValues{
						Bytes:      utils.Atofloat64(fields[1]),
						Packets:    utils.Atofloat64(fields[2]),
						Errors:     utils.Atofloat64(fields[3]),
						Dropped:    utils.Atofloat64(fields[4]),
						Fifo:       utils.Atofloat64(fields[5]),
						Frame:      utils.Atofloat64(fields[6]),
						Compressed: utils.Atofloat64(fields[7]),
						Multicast:  utils.Atofloat64(fields[8]),
					},
					Transmit: NifTransmitValues{
						Bytes:      utils.Atofloat64(fields[9]),
						Packets:    utils.Atofloat64(fields[10]),
						Errors:     utils.Atofloat64(fields[11]),
						Dropped:    utils.Atofloat64(fields[12]),
						Fifo:       utils.Atofloat64(fields[13]),
						Collisions: utils.Atofloat64(fields[14]),
						Carrier:    utils.Atofloat64(fields[15]),
						Compressed: utils.Atofloat64(fields[16]),
					},
				}

				last, exists := poller.last[device]

				if exists {
					difference := current.CalcDifference(last)

					poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "receive", "bytes"}, difference.Receive.Bytes}
					poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "receive", "packets"}, difference.Receive.Packets}
					poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "receive", "errors"}, difference.Receive.Errors}
					poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "receive", "dropped"}, difference.Receive.Dropped}
					poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "receive", "errors", "fifo"}, difference.Receive.Fifo}
					poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "receive", "errors", "frame"}, difference.Receive.Frame}
					poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "receive", "compressed"}, difference.Receive.Compressed}
					poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "receive", "multicast"}, difference.Receive.Multicast}

					poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "transmit", "bytes"}, difference.Transmit.Bytes}
					poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "transmit", "packets"}, difference.Transmit.Packets}
					poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "transmit", "errors"}, difference.Transmit.Errors}
					poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "transmit", "dropped"}, difference.Transmit.Dropped}
					poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "transmit", "errors", "fifo"}, difference.Transmit.Fifo}
					poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "transmit", "errors", "collisions"}, difference.Transmit.Collisions}
					poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "transmit", "errors", "carrier"}, difference.Transmit.Carrier}
					poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{nakedDevice, "transmit", "compressed"}, difference.Transmit.Compressed}
				}

				poller.last[device] = current
			}
		}
	}
}

func (poller NetworkInterface) Name() string {
	return "nif"
}
