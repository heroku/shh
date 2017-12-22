package shh

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/heroku/shh/Godeps/_workspace/src/github.com/heroku/slog"
)

var sampleHistogram = `
{
    "value": {
	"arithmetic_mean": 45.278688524590166,
	"geometric_mean": 29.43350463256889,
	"harmonic_mean": 10.27904811776835,
	"histogram": {
	  "28": 24,
	  "61": 14,
	  "91": 17,
	  "111": 6,
	  "141": 0
	},
	"kurtosis": -1.3219008587572214,
	"n": 61,
	"max": 99,
	"median": 43,
	"min": 1,
	"percentile": {
	  "50": 43,
	  "75": 72,
	  "95": 93,
	  "99": 99,
	  "999": 99
	},
	"skewness": 0.16028481213893223,
	"standard_deviation": 30.541300533071052,
	"variance": 932.7710382513661
  }
}
`

func TestPollHistogram(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/_metrics" {
			if r.URL.Query().Get("info") != "true" {
				t.Error("invalid query string in request: %v", r)
				http.Error(w, "unexpected request", http.StatusInternalServerError)
			}
			w.Write([]byte("{\"test\":{\"type\":\"histogram\"}}"))
			return
		}
		if r.URL.Path == "/_metrics/test" {
			w.Write(bytes.NewBufferString(sampleHistogram).Bytes())
			return
		}
		t.Errorf("unexpected request: %v", r)
		http.Error(w, "unexpected request", http.StatusInternalServerError)
	}))
	defer srv.Close()

	// a histogram produces a certain number of measurements
	measurements := make(chan Measurement, 5)
	poller := FolsomPoller{
		baseUrl:      srv.URL,
		client:       http.DefaultClient,
		measurements: measurements,
	}

	tick := time.Now()
	ctx := make(slog.Context)
	poller.doMetricsPoll(ctx, tick)

	// We expect the measurements to come in order which is not a
	// requirement but makes it easier to test.
	n := <-measurements
	if n.Name("") != "folsom.test.n" || n.StrValue() != "61" {
		t.Errorf("unexpected measurement: %v", n)
		return
	}

	max := <-measurements
	if max.Name("") != "folsom.test.max" || max.StrValue() != "99.000000" {
		t.Errorf("unexpected measurement: %v", max)
		return
	}

	median := <-measurements
	if median.Name("") != "folsom.test.median" || median.StrValue() != "43.000000" {
		t.Errorf("unexpected measurement: %v", median)
		return
	}

	p95 := <-measurements
	if p95.Name("") != "folsom.test.p95" || p95.StrValue() != "93.000000" {
		t.Errorf("unexpected measurement: %v", p95)
		return
	}

	p99 := <-measurements
	if p99.Name("") != "folsom.test.p99" || p99.StrValue() != "99.000000" {
		t.Errorf("unexpected measurement: %v", p99)
		return
	}
}
