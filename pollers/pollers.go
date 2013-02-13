package pollers

import (
	"github.com/freeformz/shh/mm"
	"github.com/freeformz/shh/utils"
	"sync"
	"time"
)

const (
	DEFAULT_POLLERS = "load,cpu,df,disk,listen,mem,nif,ntpdate,processes,self" // Default pollers
)

var (
	pollers = utils.GetEnvWithDefaultStrings("SHH_POLLERS", DEFAULT_POLLERS)
)

type Poller interface {
	Name() string
	Exit()
	Poll(tick time.Time)
}

func NewMultiPoller(measurements chan<- *mm.Measurement) Multi {
	mp := Multi{pollers: make(map[string]Poller), measurements: measurements, counts: make(map[string]uint64)}

	for _, poller := range pollers {
		switch poller {
		case "load":
			mp.RegisterPoller(NewLoadPoller(measurements))
		case "cpu":
			mp.RegisterPoller(NewCpuPoller(measurements))
		case "df":
			mp.RegisterPoller(NewDfPoller(measurements))
		case "disk":
			mp.RegisterPoller(NewDiskPoller(measurements))
		case "listen":
			mp.RegisterPoller(NewListenPoller(measurements))
		case "mem":
			mp.RegisterPoller(NewMemoryPoller(measurements))
		case "nif":
			mp.RegisterPoller(NewNetworkInterfacePoller(measurements))
		case "ntpdate":
			mp.RegisterPoller(NewNtpdatePoller(measurements))
		case "processes":
			mp.RegisterPoller(NewProcessesPoller(measurements))
		case "self":
			mp.RegisterPoller(NewSelfPoller(measurements))
		}
	}

	return mp
}

type Multi struct {
	sync.WaitGroup
	measurements chan<- *mm.Measurement
	pollers      map[string]Poller
	counts       map[string]uint64
}

func (mp Multi) RegisterPoller(poller Poller) {
	mp.pollers[poller.Name()] = poller
	mp.counts[poller.Name()] = 0
}

func (mp Multi) durationMetric(tick time.Time, name string, start time.Time) {
	mp.measurements <- &mm.Measurement{tick, mp.Name(), []string{"duration", name, "seconds"}, time.Since(start).Seconds()}
}

func (mp Multi) incrementCount(pname string) uint64 {
	count := mp.counts[pname]
	count++
	mp.counts[pname] = count
	return count
}

func (mp Multi) Poll(tick time.Time) {
	start := time.Now()
	for name, poller := range mp.pollers {
		mp.measurements <- &mm.Measurement{tick, mp.Name(), []string{"ticks", name, "count"}, mp.incrementCount(name)}
		mp.Add(1)
		go func(poller Poller) {
			start := time.Now()
			poller.Poll(tick)
			mp.Done()
			mp.durationMetric(tick, poller.Name(), start)
		}(poller)
	}
	mp.Wait()
	mp.durationMetric(tick, "all", start)
}

func (mp Multi) Name() string {
	return "multi_poller"
}

func (mp Multi) Exit() {
	for _, poller := range mp.pollers {
		poller.Exit()
	}
}
