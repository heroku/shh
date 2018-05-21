package shh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	LibratoBacklog         = 8 // No more than N pending batches in-flight
	LibratoMaxAttempts     = 4 // Max attempts before dropping batch
	LibratoStartingBackoff = 500 * time.Millisecond
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
	interval     time.Duration
	round        bool
	meta         bool
}

func NewLibratoOutputter(measurements <-chan Measurement, config Config) *Librato {
	var user string
	var token string

	if config.LibratoUrl.User != nil {
		user = config.LibratoUrl.User.Username()
		token, _ = config.LibratoUrl.User.Password()
		config.LibratoUrl.User = nil
	}

	// override settings in URL if they were present
	if config.LibratoUser != "" {
		user = config.LibratoUser
	}
	if config.LibratoToken != "" {
		token = config.LibratoToken
	}

	return &Librato{
		measurements: measurements,
		prefix:       config.Prefix,
		source:       config.Source,
		batches:      make(chan []Measurement, LibratoBacklog),
		Timeout:      config.LibratoBatchTimeout,
		BatchSize:    config.LibratoBatchSize,
		User:         user,
		Token:        token,
		Url:          config.LibratoUrl.String(),
		interval:     config.Interval,
		round:        config.LibratoRound,
		userAgent:    config.UserAgent,
		client:       &http.Client{Timeout: config.NetworkTimeout},
		meta:         config.Meta,
	}
}

func (out *Librato) Start() {
	go out.deliver()
	go out.batch()
}

// Returns a batch that is ready to be submitted to Librato, either because it timed out
// after receiving it's first measurement or it is full.
func (out *Librato) readyBatch() []Measurement {
	batch := make([]Measurement, 0, out.BatchSize)
	timer := new(time.Timer) // "empty" timer so we don't timeout before we have any measurements
	for {
		select {
		case measurement := <-out.measurements:
			batch = append(batch, measurement)
			if len(batch) == 1 { // We got a measurement, so we want to start the timer.
				timer = time.NewTimer(out.Timeout)
				defer timer.Stop()
			}
			if len(batch) == cap(batch) {
				return batch
			}
		case <-timer.C:
			return batch
		}
	}
}

// Continuously batch measurments into the batch channel
func (out *Librato) batch() {
	ctx := slog.Context{"fn": "batch", "outputter": "librato"}
	for {
		batch := out.readyBatch()

		select {
		case out.batches <- batch:
		default:
			LogError(ctx, nil, "Batches backlogged, dropping")
		}
	}
}

func (out *Librato) appendLibratoMetric(counters, gauges []LibratoMetric, mm Measurement) ([]LibratoMetric, []LibratoMetric) {
	var t int64
	attrs := LibratoMetricAttrs{UnitName: mm.Unit().Name(), UnitAbbr: mm.Unit().Abbr()}

	if out.round {
		t = mm.Time().Round(out.interval).Unix()
	} else {
		t = mm.Time().Unix()
	}

	libratoMetric := LibratoMetric{mm.Name(out.prefix), mm.Value(), t, out.source, attrs}

	switch mm.Type() {
	case CounterType:
		counters = append(counters, libratoMetric)
	case GaugeType, FloatGaugeType:
		gauges = append(gauges, libratoMetric)
	}
	return counters, gauges
}

func (out *Librato) deliver() {
	ctx := slog.Context{"fn": "prepare", "outputter": "librato"}
	for batch := range out.batches {
		gauges := make([]LibratoMetric, 0)
		counters := make([]LibratoMetric, 0)
		for _, mm := range batch {
			counters, gauges = out.appendLibratoMetric(counters, gauges, mm)
		}

		if out.meta {
			counters, gauges = out.appendLibratoMetric(
				counters,
				gauges,
				GaugeMeasurement{time.Now(), "librato-outlet", []string{"batch", "guage", "size"}, uint64(len(gauges) + 2), Metrics},
			)
			counters, gauges = out.appendLibratoMetric(
				counters,
				gauges,
				GaugeMeasurement{time.Now(), "librato-outlet", []string{"batch", "counter", "size"}, uint64(len(counters)), Metrics},
			)
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
	ctx := slog.Context{"fn": "sendWithBackoff", "outputter": "librato", "backoff": LibratoStartingBackoff, "attempts": 0}

	for ctx["attempts"].(int) < LibratoMaxAttempts {
		retry, err := out.send(payload)
		if retry {
			LogError(ctx, err, "backing off")
			ctx["backoff"] = backoff(ctx["backoff"].(time.Duration))
		} else {
			if err != nil {
				LogError(ctx, err, "error sending, no retry")
				return false
			} else {
				return true
			}
		}
		ctx["attempts"] = ctx["attempts"].(int) + 1
	}
	return false
}

// Attempts to send the payload and signals retries on errors
func (out *Librato) send(payload []byte) (bool, error) {
	body := bytes.NewReader(payload)
	req, err := http.NewRequest("POST", out.Url, body)
	if err != nil {
		return false, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", out.userAgent)
	req.SetBasicAuth(out.User, out.Token)

	resp, err := out.client.Do(req)
	if err != nil {
		return true, err
	} else {
		defer resp.Body.Close()

		if resp.StatusCode >= 300 {
			b, _ := ioutil.ReadAll(resp.Body)

			if resp.StatusCode >= 500 {
				err = fmt.Errorf("server error: %d, body: %+q", resp.StatusCode, string(b))
				return true, err
			} else {
				err = fmt.Errorf("client error: %d, body: %+q", resp.StatusCode, string(b))
				return false, err
			}

		}
	}

	return false, nil
}

// Sleeps `bo` and then returns double
func backoff(bo time.Duration) time.Duration {
	time.Sleep(bo)
	return bo * 2
}
