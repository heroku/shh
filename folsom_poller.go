package shh

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/heroku/slog"
)

type FolsomEts struct {
	Compressed bool   `json:"compressed"`
	Memory     uint64 `json:"memory"`
	Owner      string `json:"owner"`
	Heir       string `json:"heir"`
	Name       string `json:"name"`
	Size       uint64 `json:"size"`
	Node       string `json:"node"`
	NamedTable bool   `json:"named_table"`
	Type       string `json:"type"`
	KeyPos     uint64 `json:"keyos"`
	Protection string `json:"protection"`
}

type FolsomMemory struct {
	Total         uint64 `json:"total"`
	Processes     uint64 `json:"processes"`
	ProcessesUsed uint64 `json:"processes_used"`
	System        uint64 `json:"system"`
	Atom          uint64 `json:"atom"`
	AtomUsed      uint64 `json:"atom_used"`
	Binary        uint64 `json:"binary"`
	Code          uint64 `json:"code"`
	Ets           uint64 `json:"ets"`
}

type FolsomStatistics struct {
	ContextSwitches   uint64                  `json:"context_switches"`
	GarbageCollection FolsomGarbageCollection `json:"garbage_collection"`
	Io                FolsomIo                `json:"io"`
	Reductions        FolsomReductions        `json:"reductions"`
	RunQueue          uint64                  `json:"run_queue"`
	Runtime           FolsomRuntime           `json:"runtime"`
	WallClock         FolsomWallClock         `json:"wall_clock"`
}

type FolsomGarbageCollection struct {
	NumOfGcs       uint64 `json:"number_of_gcs"`
	WordsReclaimed uint64 `json:"words_reclaimed"`
}

type FolsomIo struct {
	Input  uint64 `json:"input"`
	Output uint64 `json:"output"`
}

type FolsomReductions struct {
	Total     uint64 `json:"total_reductions"`
	SinceLast uint64 `json:"reductions_since_last_call"`
}

type FolsomRuntime struct {
	Total     uint64 `json:"total_run_time"`
	SinceLast uint64 `json:"time_since_last_call"`
}

type FolsomWallClock struct {
	Total     uint64 `json:"total_wall_clock_time"`
	SinceLast uint64 `json:"wall_clock_time_since_last_call"`
}

type FolsomHistogram struct {
	ArithmeticMean    float64            `json:"arithmetic_mean"`
	GeometricMean     float64            `json:"geometric_mean"`
	HarmonicMean      float64            `json:"harmonic_mean"`
	Histogram         map[string]float64 `json:"histogram"`
	Kurtosis          float64            `json:"kurtosis"`
	N                 uint64             `json:"n"`
	Max               float64            `json:"max"`
	Median            float64            `json:"median"`
	Min               float64            `json:"min"`
	Percentile        map[string]float64 `json:"percentile"`
	Skewness          float64            `json:"skewness"`
	StandardDeviation float64            `json:"standard_deviation"`
	Variance          float64            `json:"variance"`
}

type FolsomType struct {
	Type string `json:"type"`
}

type FolsomValue struct {
	Name  string
	Type  string
	Value json.Number `json:"value"`
}

type FolsomPoller struct {
	measurements chan<- Measurement
	baseUrl      string
	client       *http.Client
}

func NewFolsomPoller(measurements chan<- Measurement, config Config) FolsomPoller {
	var url string

	if config.FolsomBaseUrl != nil {
		url = config.FolsomBaseUrl.String()
	}

	client := &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: config.NetworkTimeout,
			Dial: func(network, address string) (net.Conn, error) {
				return net.DialTimeout(network, address, config.NetworkTimeout)
			},
		},
	}

	return FolsomPoller{
		measurements: measurements,
		baseUrl:      url,
		client:       client,
	}
}

func (poller FolsomPoller) Poll(tick time.Time) {
	if poller.baseUrl == "" {
		return
	}

	ctx := slog.Context{"poller": poller.Name(), "fn": "Poll", "tick": tick}

	poller.doMemoryPoll(ctx, tick)
	poller.doStatisticsPoll(ctx, tick)
	poller.doEtsPoll(ctx, tick)
	poller.doMetricsPoll(ctx, tick)
}

func (poller FolsomPoller) doMemoryPoll(ctx slog.Context, tick time.Time) {
	memory := FolsomMemory{}

	if err := poller.decodeReq("/_memory", &memory); err != nil {
		LogError(ctx, err, "while performing request for this tick")
		return
	}

	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"mem", "total"}, memory.Total, Bytes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"mem", "procs", "total"}, memory.Processes, Bytes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"mem", "procs", "used"}, memory.ProcessesUsed, Bytes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"mem", "system"}, memory.System, Bytes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"mem", "atom", "total"}, memory.Atom, Bytes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"mem", "atom", "used"}, memory.AtomUsed, Bytes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"mem", "binary"}, memory.Binary, Bytes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"mem", "code"}, memory.Code, Bytes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"mem", "ets"}, memory.Ets, Bytes}
}

func (poller FolsomPoller) doStatisticsPoll(ctx slog.Context, tick time.Time) {
	stats := FolsomStatistics{}
	if err := poller.decodeReq("/_statistics", &stats); err != nil {
		LogError(ctx, err, "while performing request for this tick")
		return
	}

	poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{"stats", "context-switches"}, stats.ContextSwitches, ContextSwitches}
	poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{"stats", "gc", "num"}, stats.GarbageCollection.NumOfGcs, Empty}
	poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{"stats", "gc", "reclaimed"}, stats.GarbageCollection.WordsReclaimed, Words}
	poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{"stats", "io", "input"}, stats.Io.Input, Bytes}
	poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{"stats", "io", "output"}, stats.Io.Output, Bytes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"stats", "reductions"}, stats.Reductions.SinceLast, Reductions}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"stats", "run-queue"}, stats.RunQueue, Processes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"stats", "runtime"}, stats.Runtime.SinceLast, MilliSeconds}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"stats", "wall-clock"}, stats.WallClock.SinceLast, MilliSeconds}
}
func (poller FolsomPoller) doEtsPoll(ctx slog.Context, tick time.Time) {
	tables := make(map[string]FolsomEts)

	if err := poller.decodeReq("/_ets", &tables); err != nil {
		LogError(ctx, err, "while performing request for this tick")
		return
	}

	for _, tab := range tables {
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"ets", tab.Name, "memory"}, tab.Memory, Words}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"ets", tab.Name, "size"}, tab.Size, Terms}
	}
}

func (poller FolsomPoller) doMetricsPoll(ctx slog.Context, tick time.Time) {
	metrics := make(map[string]FolsomType)
	if err := poller.decodeReq("/_metrics?info=true", &metrics); err != nil {
		LogError(ctx, err, "while performing request for this tick")
		return
	}

	for key, ft := range metrics {
		switch ft.Type {
		case "counter", "gauge":
			v := FolsomValue{Name: key, Type: ft.Type}
			if err := poller.decodeReq("/_metrics/"+v.Name, &v); err != nil {
				LogError(ctx, err, "while performing request for "+v.Name+" this tick")
				return
			}

			if m, err := poller.genMeasurement(tick, v); err != nil {
				LogError(ctx, err, "while performing request for "+v.Name+" this tick")
				return
			} else {
				poller.measurements <- m
			}
		case "histogram":
			v := struct {
				Value FolsomHistogram `json:"value"`
			}{}

			if err := poller.decodeReq("/_metrics/"+key, &v); err != nil {
				LogError(ctx, err, "while performing request for "+key+" this tick")
				return
			}

			if err := poller.genHistogram(tick, key, v.Value); err != nil {
				LogError(ctx, err, "while performing request for "+key+" this tick")
				return
			}
		default:
			LogError(ctx, errors.New("Unsupported metric type: "+ft.Type), "while performing request for "+key+" this tick")
			return
		}
	}
}

func (poller FolsomPoller) genMeasurement(tick time.Time, v FolsomValue) (Measurement, error) {
	var err error

	switch v.Type {
	case "counter":
		var val int64
		if val, err = v.Value.Int64(); err == nil {
			return CounterMeasurement{tick, poller.Name(), []string{v.Name}, uint64(val), Empty}, nil
		}
	case "gauge":
		if strings.Contains(v.Value.String(), ".") {
			var val float64
			if val, err = v.Value.Float64(); err == nil {
				return FloatGaugeMeasurement{tick, poller.Name(), []string{v.Name}, val, Empty}, nil
			}
		} else {
			var val int64
			if val, err = v.Value.Int64(); err == nil {
				return GaugeMeasurement{tick, poller.Name(), []string{v.Name}, uint64(val), Empty}, nil
			}
		}
	default:
		err = errors.New("Unsupported metric type: " + v.Type)
	}

	return nil, err
}

func (poller FolsomPoller) genHistogram(tick time.Time, name string, histogram FolsomHistogram) error {
	// number of samples in histogram
	n := GaugeMeasurement{tick, poller.Name(), []string{name, "n"}, histogram.N, Empty}
	poller.measurements <- n

	max := FloatGaugeMeasurement{tick, poller.Name(), []string{name, "max"}, histogram.Max, Empty}
	poller.measurements <- max

	median := FloatGaugeMeasurement{tick, poller.Name(), []string{name, "median"}, histogram.Median, Empty}
	poller.measurements <- median

	if v, ok := histogram.Percentile["95"]; !ok {
		return errors.New("failed to extract p95 from histogram")
	} else {
		p95 := FloatGaugeMeasurement{tick, poller.Name(), []string{name, "p95"}, v, Empty}
		poller.measurements <- p95
	}

	if v, ok := histogram.Percentile["99"]; !ok {
		return errors.New("failed to extract p99 from histogram")
	} else {
		p99 := FloatGaugeMeasurement{tick, poller.Name(), []string{name, "p99"}, v, Empty}
		poller.measurements <- p99
	}

	return nil
}

func (poller FolsomPoller) decodeReq(path string, v interface{}) error {
	req, err := http.NewRequest("GET", poller.baseUrl+path, nil)
	if err != nil {
		return err
	}

	resp, rerr := poller.client.Do(req)
	if rerr != nil {
		return rerr
	} else if resp.StatusCode >= 300 {
		resp.Body.Close()
		return fmt.Errorf("Response returned a %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	if derr := decoder.Decode(v); derr != nil {
		return derr
	}

	return nil
}

func (poller FolsomPoller) Name() string {
	return "folsom"
}

func (poller FolsomPoller) Exit() {}
