package main

import (
	"time"
)

const (
	DEVICE_FILE = "/proc/net/dev"
)

type NetworkInterface struct {
	measurements chan<- Measurement
	Devices      []string
}

func NewNetworkInterfacePoller(measurements chan<- Measurement, config Config) NetworkInterface {
	return NetworkInterface{
		measurements: measurements,
		Devices:      config.NifDevices,
	}
}

// http://www.kernel.org/doc/Documentation/filesystems/proc.txt (section 1.4)
func (poller NetworkInterface) Poll(tick time.Time) {

	for line := range FileLineChannel(DEVICE_FILE) {
		fields := Fields(line)
		device := fields[0]

		if SliceContainsString(poller.Devices, device) {
			// It's a device we want to gather metrics for

			poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{device, "receive", "bytes"}, Atouint64(fields[1])}
			poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{device, "receive", "packets"}, Atouint64(fields[2])}
			poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{device, "receive", "errors"}, Atouint64(fields[3])}
			poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{device, "receive", "dropped"}, Atouint64(fields[4])}
			poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{device, "receive", "errors", "fifo"}, Atouint64(fields[5])}
			poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{device, "receive", "errors", "frame"}, Atouint64(fields[6])}
			poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{device, "receive", "compressed"}, Atouint64(fields[7])}
			poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{device, "receive", "multicast"}, Atouint64(fields[8])}

			poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{device, "transmit", "bytes"}, Atouint64(fields[9])}
			poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{device, "transmit", "packets"}, Atouint64(fields[10])}
			poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{device, "transmit", "errors"}, Atouint64(fields[11])}
			poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{device, "transmit", "dropped"}, Atouint64(fields[12])}
			poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{device, "transmit", "errors", "fifo"}, Atouint64(fields[13])}
			poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{device, "transmit", "errors", "collisions"}, Atouint64(fields[14])}
			poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{device, "transmit", "errors", "carrier"}, Atouint64(fields[15])}
			poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{device, "transmit", "compressed"}, Atouint64(fields[16])}

		}
	}
}

func (poller NetworkInterface) Name() string {
	return "nif"
}

func (poller NetworkInterface) Exit() {}
