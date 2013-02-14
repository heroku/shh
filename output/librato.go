package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/freeformz/shh/config"
	"github.com/freeformz/shh/mm"
	"io/ioutil"
	"log"
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
	batches chan []*mm.Measurement = make(chan []*mm.Measurement, 4)
)

type Librato struct {
	measurements <-chan *mm.Measurement
}

func NewLibratoOutputter(measurements <-chan *mm.Measurement) Librato {
	return Librato{measurements: measurements}
}

func (out Librato) Start() {
	go out.deliver()
	go out.batch()
}

func (out Librato) batch() {
	ticker := time.Tick(config.LibratoBatchTimeout)
	batch := makeBatch()
	for {
		select {
		case measurement := <-out.measurements:
			batch = append(batch, measurement)
			if len(batch) == cap(batch) {
				batches <- batch
				batch = makeBatch()
			}
		case <-ticker:
			if len(batch) > 0 {
				batches <- batch
				batch = makeBatch()
			}
		}
	}
}

func makeBatch() []*mm.Measurement {
	return make([]*mm.Measurement, 0, config.LibratoBatchSize)
}

func (out Librato) deliver() {
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
			log.Fatal(err)
		}

		body := bytes.NewBuffer(j)
		req, err := http.NewRequest("POST", LIBRATO_URL, body)
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Add("Content-Type", "application/json")
		req.SetBasicAuth(config.LibratoUser, config.LibratoToken)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		if resp.StatusCode/100 != 2 {
			b, _ := ioutil.ReadAll(resp.Body)
			fmt.Printf("%s\n", b)
		}
		resp.Body.Close()
	}
}
