package shh

import (
	"sync"
	"time"
)

type Poller interface {
	Name() string
	Exit()
	Poll(tick time.Time)
}

func NewMultiPoller(measurements chan<- Measurement, config Config) Multi {
	mp := Multi{pollers: make(map[string]Poller), measurements: measurements}

	for _, poller := range config.Pollers {
		switch poller {
		case "load":
			mp.RegisterPoller(NewLoadPoller(measurements))
		case "cpu":
			mp.RegisterPoller(NewCpuPoller(measurements, config))
		case "df":
			mp.RegisterPoller(NewDfPoller(measurements, config))
		case "disk":
			mp.RegisterPoller(NewDiskPoller(measurements, config))
		case "filenr":
			mp.RegisterPoller(NewFileNrPoller(measurements))
		case "listen":
			mp.RegisterPoller(NewListenPoller(measurements, config))
		case "mem":
			mp.RegisterPoller(NewMemoryPoller(measurements, config))
		case "nagios3stats":
			mp.RegisterPoller(NewNagios3StatsPoller(measurements, config))
		case "nif":
			mp.RegisterPoller(NewNetworkInterfacePoller(measurements, config))
		case "ntpdate":
			mp.RegisterPoller(NewNtpdatePoller(measurements, config))
		case "processes":
			mp.RegisterPoller(NewProcessesPoller(measurements, config))
		case "self":
			mp.RegisterPoller(NewSelfPoller(measurements, config))
		case "conntrack":
			mp.RegisterPoller(NewConntrackPoller(measurements))
		case "syslogngstats":
			mp.RegisterPoller(NewSyslogngStatsPoller(measurements, config))
		case "sockstat":
			mp.RegisterPoller(NewSockStatPoller(measurements, config))
		case "splunksearchpeers":
			mp.RegisterPoller(NewSplunkSearchPeersPoller(measurements, config))
		case "folsom":
			mp.RegisterPoller(NewFolsomPoller(measurements, config))
		}
	}

	return mp
}

type Multi struct {
	sync.WaitGroup
	measurements chan<- Measurement
	pollers      map[string]Poller
}

func (mp Multi) RegisterPoller(poller Poller) {
	mp.pollers[poller.Name()] = poller
}

func (mp Multi) durationMetric(tick time.Time, name string, start time.Time) {
	mp.measurements <- FloatGaugeMeasurement{tick, mp.Name(), []string{"duration", name, "seconds"}, time.Since(start).Seconds(), Seconds}
}

func (mp Multi) Poll(tick time.Time) {
	defer mp.durationMetric(tick, "all", time.Now())

	for _, poller := range mp.pollers {
		mp.Add(1)
		go func(poller Poller) {
			defer mp.durationMetric(tick, poller.Name(), time.Now())
			poller.Poll(tick)
			mp.Done()
		}(poller)
	}

	mp.Wait()
}

func (mp Multi) Name() string {
	return "multi_poller"
}

func (mp Multi) Exit() {
	for _, poller := range mp.pollers {
		poller.Exit()
	}
}
