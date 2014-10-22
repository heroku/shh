package shh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/heroku/slog"
)

type LibratoMetric struct {
	Name       string             `json:"name"`
	Value      interface{}        `json:"value"`
	When       int64              `json:"measure_time"`
	Source     string             `json:"source,omitempty"`
	Attributes LibratoMetricAttrs `json:"attributes,omitempty"`
}

type LibratoMetricAttrs struct {
	UnitName string `json:"display_units_long,omitempty"`
	UnitAbbr string `json:"display_units_short,omitempty"`
}

type LibratoPostBody struct {
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
	userAgent    string
}

func NewLibratoOutputter(measurements <-chan Measurement, config Config) *Librato {
	return &Librato{
		measurements: measurements,
		prefix:       config.Prefix,
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
		userAgent: config.UserAgent,
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
	ctx := slog.Context{"fn": "batch", "outputter": "librato"}
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
				LogError(ctx, nil, "Batches backlogged, dropping")
			}
			batch = out.makeBatch()
			ready = false
		}
	}
}

func (out *Librato) deliver() {
	ctx := slog.Context{"fn": "prepare", "outputter": "librato"}
	for batch := range out.batches {
		gauges := make([]LibratoMetric, 0)
		counters := make([]LibratoMetric, 0)
		for _, mm := range batch {
			attrs := LibratoMetricAttrs{UnitName: mm.Unit().Name(), UnitAbbr: mm.Unit().Abbr()}
			libratoMetric := LibratoMetric{mm.Name(out.prefix), mm.Value(), mm.Time().Unix(), out.source, attrs}
			switch mm.Type() {
			case CounterType:
				counters = append(counters, libratoMetric)
			case GaugeType, FloatGaugeType:
				gauges = append(gauges, libratoMetric)
			}
		}

		payload := LibratoPostBody{gauges, counters}
		j, err := json.Marshal(payload)
		if err != nil {
			FatalError(ctx, err, "marshaling json")
		}

		out.sendWithBackoff(j)
	}
}

func (out *Librato) sendWithBackoff(payload []byte) bool {
	ctx := slog.Context{"fn": "retry", "outputter": "librato"}
	attempts := 0
	bo := 0 * time.Millisecond

	for attempts < LibratoMaxAttempts {
		retry, err := out.send(ctx, payload)
		if retry {
			LogError(ctx, err, "backoffing off")
			bo = backoff(bo)
		} else if err != nil {
			LogError(ctx, err, "error sending")
			return false
		} else {
			return true
		}

		resetCtx(ctx)
		attempts++
	}
	return false
}

func (out *Librato) send(ctx slog.Context, payload []byte) (retry bool, e error) {
	body := bytes.NewBuffer(payload)
	req, err := http.NewRequest("POST", out.Url, body)
	if err != nil {
		FatalError(ctx, err, "creating new request")
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", out.userAgent)
	req.SetBasicAuth(out.User, out.Token)

	resp, err := out.client.Do(req)
	if err != nil {
		if nerr, ok := err.(net.Error); ok && (nerr.Temporary() || nerr.Timeout()) {
			retry = true
			e = fmt.Errorf("Backing off due to transport error")
		} else if strings.Contains(err.Error(), "timeout awaiting response") {
			retry = false
			e = err
		} else if err == io.EOF {
			retry = true
			e = fmt.Errorf("Backing off due to EOF")
		} else {
			FatalError(ctx, err, "doing request")
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

func resetCtx(ctx slog.Context) {
	delete(ctx, "body")
	delete(ctx, "code")
}
