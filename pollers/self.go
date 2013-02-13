package pollers

import (
	"github.com/freeformz/shh/mm"
	"runtime"
	"time"
)

type Self struct {
	measurements chan<- *mm.Measurement
	stats        runtime.MemStats
}

func NewSelfPoller(measurements chan<- *mm.Measurement) Self {
	return Self{measurements: measurements, stats: runtime.MemStats{}}
}

// See http://golang.org/pkg/runtime/#MemStats
func (poller Self) Poll(tick time.Time) {
	runtime.ReadMemStats(&poller.stats)

	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"measurements", "length"}, float64(len(poller.measurements))}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "general", "alloc", "inuse", "bytes"}, float64(poller.stats.Alloc)} // GUAGE
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "general", "alloc", "bytes"}, poller.stats.TotalAlloc}              // COUNTER
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "general", "sys", "bytes"}, float64(poller.stats.Sys)}              // GUAGE
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "general", "pointer", "lookups"}, poller.stats.Lookups}             // COUNTER
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "general", "mallocs"}, poller.stats.Mallocs}                        // COUNTER
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "general", "frees"}, poller.stats.Frees}                            // COUNTER

	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "heap", "alloc", "bytes"}, float64(poller.stats.HeapAlloc)}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "heap", "sys", "bytes"}, float64(poller.stats.HeapSys)}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "heap", "idle", "bytes"}, float64(poller.stats.HeapIdle)}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "heap", "inuse", "bytes"}, float64(poller.stats.HeapInuse)}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "heap", "released", "bytes"}, float64(poller.stats.HeapReleased)}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "heap", "objects"}, float64(poller.stats.HeapObjects)}

	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "stack", "inuse"}, float64(poller.stats.StackInuse)}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "stack", "sys"}, float64(poller.stats.StackSys)}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "mspan", "inuse"}, float64(poller.stats.MSpanInuse)}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "mspan", "sys"}, float64(poller.stats.MSpanSys)}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "mcache", "inuse"}, float64(poller.stats.MCacheInuse)}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "mcache", "sys"}, float64(poller.stats.MCacheSys)}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "buckhash", "sys"}, float64(poller.stats.BuckHashSys)}

	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "gc", "next"}, float64(poller.stats.NextGC)}
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "gc", "pause", "ns"}, poller.stats.PauseTotalNs} // COUNTER
	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "gc", "num"}, uint64(poller.stats.NumGC)}        // COUNTER

	poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{"memstats", "goroutines", "num"}, float64(runtime.NumGoroutine())}
}

func (poller Self) Name() string {
	return "self"
}

func (poller Self) Exit() {}
