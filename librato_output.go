package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

type LibratoMetric struct {
	Name   string      `json:"name"`
	Value  interface{} `json:"value"`
	When   int64       `json:"measure_time"`
	Source string      `json:"source,omitempty"`
}

type PostBody struct {
	Gauges   []LibratoMetric `json:"gauges,omitempty"`
	Counters []LibratoMetric `json:"counters,omitempty"`
}

const (
	LibratoBacklog               = 8 // No more than N pending batches in-flight
	LibratoMaxAttempts           = 4 // Max attempts before dropping batch
	LibratoStartingBackoffMillis = 200 * time.Millisecond
)

type Librato struct {
	Timeout      time.Duration
	BatchSize    int
	User         string
	Token        string
	Url          string
	measurements <-chan Measurement
	batches      chan []Measurement
	prefix       string
	source       string
	client       *http.Client
}

func NewLibratoOutputter(measurements <-chan Measurement, config Config) *Librato {
	return &Librato{
		measurements: measurements,
		source:       config.Source,
		batches:      make(chan []Measurement, LibratoBacklog),
		Timeout:      config.LibratoBatchTimeout,
		BatchSize:    config.LibratoBatchSize,
		User:         config.LibratoUser,
		Token:        config.LibratoToken,
		Url:          config.LibratoUrl,
		client: &http.Client{
			Transport: &http.Transport{
				ResponseHeaderTimeout: config.LibratoNetworkTimeout,
				Dial: func(network, address string) (net.Conn, error) {
					return net.DialTimeout(network, address, config.LibratoNetworkTimeout)
				},
			},
		},
	}
}

func (out *Librato) Start() {
	go out.deliver()
	go out.batch()
}

func (out *Librato) makeBatch() []Measurement {
	return make([]Measurement, 0, out.BatchSize)
}

func (out *Librato) batch() {
	var ready bool
	ctx := Slog{"fn": "batch", "outputter": "librato"}
	ticker := time.Tick(out.Timeout)
	batch := out.makeBatch()
	for {
		select {
		case measurement := <-out.measurements:
			batch = append(batch, measurement)
			if len(batch) == cap(batch) {
				ready = true
			}
		case <-ticker:
			if len(batch) > 0 {
				ready = true
			}
		}

		if ready {
			select {
			case out.batches <- batch:
			default:
				ctx.Error(nil, "Batches backlogged, dropping")
			}
			batch = out.makeBatch()
			ready = false
		}
	}
}

func (out *Librato) deliver() {
	ctx := Slog{"fn": "prepare", "outputter": "librato"}
	for batch := range out.batches {
		gauges := make([]LibratoMetric, 0)
		counters := make([]LibratoMetric, 0)
		for _, mm := range batch {
			libratoMetric := LibratoMetric{mm.Name(out.prefix), mm.Value(), mm.Time().Unix(), out.source}
			switch mm.Type() {
			case CounterType:
				counters = append(counters, libratoMetric)
			case GaugeType, FloatGaugeType:
				gauges = append(gauges, libratoMetric)
			}
		}

		payload := PostBody{gauges, counters}
		j, err := json.Marshal(payload)
		if err != nil {
			ctx.FatalError(err, "marshaling json")
		}

		out.sendWithBackoff(j)
	}
}

func (out *Librato) sendWithBackoff(payload []byte) bool {
	ctx := Slog{"fn": "retry", "outputter": "librato"}
	attempts := 0
	bo := 0 * time.Millisecond

	for attempts < LibratoMaxAttempts {
		if retry, err := out.send(ctx, payload); retry {
			ctx["backoff"] = bo
			ctx["message"] = err
			fmt.Println(ctx)
			bo = backoff(bo)
		} else if err != nil {
			ctx["error"] = err
			fmt.Println(ctx)
			return false
		} else {
			return true
		}

		resetCtx(ctx)
		attempts++
	}
	return false
}

func (out *Librato) send(ctx Slog, payload []byte) (retry bool, e error) {
	body := bytes.NewBuffer(payload)
	req, err := http.NewRequest("POST", out.Url, body)
	if err != nil {
		ctx.FatalError(err, "creating new request")
	}

	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(out.User, out.Token)

	resp, err := out.client.Do(req)
	if err != nil {
		if nerr, ok := err.(net.Error); ok && (nerr.Temporary() || nerr.Timeout()) {
			retry = true
			e = fmt.Errorf("Backing off due to transport error")
		} else if strings.Contains(err.Error(), "timeout awaiting response") {
			retry = false
			e = nil
		} else {
			ctx.FatalError(err, "doing request")
		}
	} else {
		defer resp.Body.Close()

		if resp.StatusCode >= 300 {
			b, _ := ioutil.ReadAll(resp.Body)
			ctx["body"] = string(b)
			ctx["code"] = resp.StatusCode

			if resp.StatusCode >= 500 {
				retry = true
				e = fmt.Errorf("Backing off due to server error")
			} else {
				e = fmt.Errorf("Client error")
			}
		}
	}

	return
}

// Sleeps `bo` and then returns double, unless 0, in which case
// returns the initial starting sleep time.
func backoff(bo time.Duration) time.Duration {
	if bo > 0 {
		time.Sleep(bo)
		return bo * 2
	} else {
		return LibratoStartingBackoffMillis
	}
}

func resetCtx(ctx Slog) {
	delete(ctx, "backoff")
	delete(ctx, "message")
	delete(ctx, "body")
	delete(ctx, "code")
}
