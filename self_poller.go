package shh

import (
	"runtime"
	"time"
)

type Self struct {
	measurements chan<- Measurement
	stats        runtime.MemStats
	full         bool
}

func NewSelfPoller(measurements chan<- Measurement, config Config) Self {
	return Self{measurements: measurements, stats: runtime.MemStats{}, full: config.SelfPollerMode == "full"}
}

// See http://golang.org/pkg/runtime/#MemStats
func (poller Self) Poll(tick time.Time) {
	runtime.ReadMemStats(&poller.stats)

	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"memstats", "goroutines", "num"}, uint64(runtime.NumGoroutine()), Routines}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"memstats", "general", "alloc", "inuse", "bytes"}, poller.stats.Alloc, Bytes}
	poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{"memstats", "general", "alloc", "bytes"}, poller.stats.TotalAlloc, Bytes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"memstats", "heap", "alloc", "bytes"}, poller.stats.HeapAlloc, Bytes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"memstats", "heap", "inuse", "bytes"}, poller.stats.HeapInuse, Bytes}

	if poller.full {
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"measurements", "length"}, uint64(len(poller.measurements)), Empty}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"memstats", "general", "sys", "bytes"}, poller.stats.Sys, Bytes}
		poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{"memstats", "general", "pointer", "lookups"}, poller.stats.Lookups, Empty}
		poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{"memstats", "general", "mallocs"}, poller.stats.Mallocs, Empty}
		poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{"memstats", "general", "frees"}, poller.stats.Frees, Empty}

		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"memstats", "heap", "sys", "bytes"}, poller.stats.HeapSys, Bytes}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"memstats", "heap", "idle", "bytes"}, poller.stats.HeapIdle, Bytes}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"memstats", "heap", "released", "bytes"}, poller.stats.HeapReleased, Bytes}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"memstats", "heap", "objects"}, poller.stats.HeapObjects, Objects}

		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"memstats", "stack", "inuse"}, poller.stats.StackInuse, Bytes}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"memstats", "stack", "sys"}, poller.stats.StackSys, Bytes}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"memstats", "mspan", "inuse"}, poller.stats.MSpanInuse, Empty}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"memstats", "mspan", "sys"}, poller.stats.MSpanSys, Empty}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"memstats", "mcache", "inuse"}, poller.stats.MCacheInuse, Empty}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"memstats", "mcache", "sys"}, poller.stats.MCacheSys, Empty}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"memstats", "buckhash", "sys"}, poller.stats.BuckHashSys, Empty}

		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"memstats", "gc", "next"}, poller.stats.NextGC, Bytes}
		poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{"memstats", "gc", "pause", "ns"}, poller.stats.PauseTotalNs, NanoSeconds}
		poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{"memstats", "gc", "num"}, uint64(poller.stats.NumGC), Empty}
	}
}

func (poller Self) Name() string {
	return "self"
}

func (poller Self) Exit() {}
