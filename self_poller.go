package main

import (
	"runtime"
	"time"
)

type Self struct {
	measurements chan<- Measurement
	stats        runtime.MemStats
}

func NewSelfPoller(measurements chan<- Measurement) Self {
	return Self{measurements: measurements, stats: runtime.MemStats{}}
}

// See http://golang.org/pkg/runtime/#MemStats
func (poller Self) Poll(tick time.Time) {
	runtime.ReadMemStats(&poller.stats)

	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"measurements", "length"}, uint64(len(poller.measurements))}
	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"memstats", "general", "alloc", "inuse", "bytes"}, poller.stats.Alloc}
	poller.measurements <- &CounterMeasurement{tick, poller.Name(), []string{"memstats", "general", "alloc", "bytes"}, poller.stats.TotalAlloc}
	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"memstats", "general", "sys", "bytes"}, poller.stats.Sys}
	poller.measurements <- &CounterMeasurement{tick, poller.Name(), []string{"memstats", "general", "pointer", "lookups"}, poller.stats.Lookups}
	poller.measurements <- &CounterMeasurement{tick, poller.Name(), []string{"memstats", "general", "mallocs"}, poller.stats.Mallocs}
	poller.measurements <- &CounterMeasurement{tick, poller.Name(), []string{"memstats", "general", "frees"}, poller.stats.Frees}

	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"memstats", "heap", "alloc", "bytes"}, poller.stats.HeapAlloc}
	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"memstats", "heap", "sys", "bytes"}, poller.stats.HeapSys}
	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"memstats", "heap", "idle", "bytes"}, poller.stats.HeapIdle}
	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"memstats", "heap", "inuse", "bytes"}, poller.stats.HeapInuse}
	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"memstats", "heap", "released", "bytes"}, poller.stats.HeapReleased}
	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"memstats", "heap", "objects"}, poller.stats.HeapObjects}

	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"memstats", "stack", "inuse"}, poller.stats.StackInuse}
	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"memstats", "stack", "sys"}, poller.stats.StackSys}
	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"memstats", "mspan", "inuse"}, poller.stats.MSpanInuse}
	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"memstats", "mspan", "sys"}, poller.stats.MSpanSys}
	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"memstats", "mcache", "inuse"}, poller.stats.MCacheInuse}
	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"memstats", "mcache", "sys"}, poller.stats.MCacheSys}
	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"memstats", "buckhash", "sys"}, poller.stats.BuckHashSys}

	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"memstats", "gc", "next"}, poller.stats.NextGC}
	poller.measurements <- &CounterMeasurement{tick, poller.Name(), []string{"memstats", "gc", "pause", "ns"}, poller.stats.PauseTotalNs}
	poller.measurements <- &CounterMeasurement{tick, poller.Name(), []string{"memstats", "gc", "num"}, uint64(poller.stats.NumGC)}

	poller.measurements <- &GaugeMeasurement{tick, poller.Name(), []string{"memstats", "goroutines", "num"}, uint64(runtime.NumGoroutine())}
}

func (poller Self) Name() string {
	return "self"
}

func (poller Self) Exit() {}
