package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	LIBRATO_URL = "https://metrics-api.librato.com/v1/metrics"
)

var (
	batches chan []*Measurement = make(chan []*Measurement, 4)
)

type Librato struct {
	measurements <-chan *Measurement
	Timeout      time.Duration
	BatchSize    int
	User         string
	Token        string
}

func NewLibratoOutputter(measurements <-chan *Measurement, config Config) Librato {
	return Librato{
		measurements: measurements,
		Timeout:      config.LibratoBatchTimeout,
		BatchSize:    config.LibratoBatchSize,
		User:         config.LibratoUser,
		Token:        config.LibratoToken,
	}
}

func (out Librato) Start() {
	go out.deliver()
	go out.batch()
}

func (out Librato) batch() {
	ticker := time.Tick(out.Timeout)
	batch := out.makeBatch()
	for {
		select {
		case measurement := <-out.measurements:
			batch = append(batch, measurement)
			if len(batch) == cap(batch) {
				batches <- batch
				batch = out.makeBatch()
			}
		case <-ticker:
			if len(batch) > 0 {
				batches <- batch
				batch = out.makeBatch()
			}
		}
	}
}

func (out Librato) makeBatch() []*Measurement {
	return make([]*Measurement, 0, out.BatchSize)
}

func (out Librato) deliver() {
	ctx := Slog{"fn": "deliver", "outputter": "librato"}

	for batch := range batches {
		gauges := make([]LibratoMetric, 0, len(batch))
		counters := make([]LibratoMetric, 0, len(batch))
		for _, metric := range batch {
			libratoMetric := LibratoMetric{metric.Measured(), metric.Value, metric.When.Unix(), metric.Source()}
			switch metric.Value.(type) {
			case uint64:
				counters = append(counters, libratoMetric)
			case float64:
				gauges = append(gauges, libratoMetric)
			}
		}

		payload := PostBody{gauges, counters}

		j, err := json.Marshal(payload)
		if err != nil {
			ctx.FatalError(err, "marshaling json")
		}

		body := bytes.NewBuffer(j)
		req, err := http.NewRequest("POST", LIBRATO_URL, body)
		if err != nil {
			ctx.FatalError(err, "creating new request")
		}

		req.Header.Add("Content-Type", "application/json")
		req.SetBasicAuth(out.User, out.Token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			ctx.FatalError(err, "doing request")
		}

		if resp.StatusCode/100 != 2 {
			b, _ := ioutil.ReadAll(resp.Body)
			ctx["body"] = b
			ctx["code"] = resp.StatusCode
			fmt.Println(ctx)
			delete(ctx, "body")
			delete(ctx, "code")
		}
		resp.Body.Close()
	}
}
