package shh

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"
)

type HappyHandler struct {
	headers http.Header
}

func (s *HappyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.headers = req.Header
	w.WriteHeader(http.StatusOK)
}

type SleepyHandler struct {
	Amt     time.Duration
	ReqIncr time.Duration
	times   int
}

func (s *SleepyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.times++
	time.Sleep(s.Amt)
	w.WriteHeader(http.StatusOK)
	s.Amt += s.ReqIncr
	if s.Amt < 0 {
		s.Amt = 0
	}
}

type GrumpyHandler struct {
	ResponseCodes []int
	idx           int
}

func (g *GrumpyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if len(g.ResponseCodes) > 0 {
		w.WriteHeader(g.ResponseCodes[g.idx])
		g.idx = (g.idx + 1) % len(g.ResponseCodes)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func TestLibrato_TimeToHeaderTimeout(t *testing.T) {
	handler := &SleepyHandler{
		Amt:     2 * time.Second,
		ReqIncr: -600 * time.Millisecond,
	}
	server := httptest.NewServer(handler)
	defer server.Close()

	config := GetConfig()
	config.LibratoUrl, _ = url.Parse(server.URL)
	config.NetworkTimeout = 1 * time.Second
	config.LibratoUser = "user"
	config.LibratoToken = "token"

	measurements := make(chan Measurement, 10)
	librato := NewLibratoOutputter(measurements, config)

	if !librato.sendWithBackoff([]byte(`{}`)) {
		t.Errorf("Request should not have errored with a sleepy handler")
	}

	if handler.times != 3 {
		t.Errorf("Request should have been tried 3 times, instead it was tried: ", handler.times)
	}
}

func TestLibrato_ServerErrorBackoff(t *testing.T) {
	handler := &GrumpyHandler{ResponseCodes: []int{503, 500, 200}}
	server := httptest.NewServer(handler)
	defer server.Close()

	config := GetConfig()
	config.LibratoUrl, _ = url.Parse(server.URL)
	config.LibratoUser = "user"
	config.LibratoToken = "token"

	measurements := make(chan Measurement, 10)
	librato := NewLibratoOutputter(measurements, config)

	if !librato.sendWithBackoff([]byte(`{}`)) {
		t.Errorf("Request should have completed successfully with a grumpy handler")
	}
}

func TestLibrato_IndefiniteBackoff(t *testing.T) {
	handler := &GrumpyHandler{ResponseCodes: []int{500}}
	server := httptest.NewServer(handler)
	defer server.Close()

	config := GetConfig()
	config.LibratoUrl, _ = url.Parse(server.URL)
	config.LibratoUser = "user"
	config.LibratoToken = "token"

	measurements := make(chan Measurement, 10)
	librato := NewLibratoOutputter(measurements, config)

	if librato.sendWithBackoff([]byte(`{}`)) {
		t.Errorf("Retry should have given up. This is an especially grumpy handler")
	}
}

func TestLibrato_ClientError(t *testing.T) {
	handler := &GrumpyHandler{ResponseCodes: []int{401}}
	server := httptest.NewServer(handler)
	defer server.Close()

	config := GetConfig()
	config.LibratoUrl, _ = url.Parse(server.URL)
	config.LibratoUser = "user"
	config.LibratoToken = "token"

	measurements := make(chan Measurement, 10)
	librato := NewLibratoOutputter(measurements, config)

	if librato.sendWithBackoff([]byte(`{}`)) {
		t.Errorf("Retry should not have succeeded due to non-server error.")
	}
}

func TestLibrato_UserAgent(t *testing.T) {
	handler := &HappyHandler{}
	server := httptest.NewServer(handler)
	defer server.Close()

	config := GetConfig()
	config.LibratoUrl, _ = url.Parse(server.URL)
	config.LibratoUser = "user"
	config.LibratoToken = "token"

	measurements := make(chan Measurement, 10)
	librato := NewLibratoOutputter(measurements, config)

	if !librato.sendWithBackoff([]byte(`{}`)) {
		t.Errorf("should have succeeded.")
	}

	h, ok := handler.headers["User-Agent"]
	if !ok {
		t.Errorf("Missing User-Agent Header")
	}

	if h[0] != config.UserAgent {
		t.Errorf("Incorrect User-Agent Header value")
	}
}

func TestLibrato_UserPassFromEnv(t *testing.T) {
	os.Setenv("SHH_LIBRATO_USER", "foo")
	os.Setenv("SHH_LIBRATO_TOKEN", "bar")
	os.Setenv("SHH_LIBRATO_URL", "http://baz:quux@librato.com")

	config := GetConfig()

	measurements := make(chan Measurement, 10)
	librato := NewLibratoOutputter(measurements, config)

	if librato.Url != "http://librato.com" {
		t.Errorf("Incorrect url for librato. Found: '%s', expected: '%s'", librato.Url, "http://librato.com")
	}

	if librato.User != "foo" {
		t.Errorf("Incorrect user for librato. Found: '%s', expected: '%s'", librato.User, "foo")
	}

	if librato.Token != "bar" {
		t.Errorf("Incorrect token for librato. Found: '%s', expected: '%s'", librato.Token, "bar")
	}
}

func TestLibrato_UserPassFromURL(t *testing.T) {
	os.Setenv("SHH_LIBRATO_USER", "")
	os.Setenv("SHH_LIBRATO_TOKEN", "")
	os.Setenv("SHH_LIBRATO_URL", "http://baz:quux@librato.com")

	config := GetConfig()

	measurements := make(chan Measurement, 10)
	librato := NewLibratoOutputter(measurements, config)

	if librato.Url != "http://librato.com" {
		t.Errorf("Incorrect url for librato. Found: '%s', expected: '%s'", librato.Url, "http://librato.com")
	}

	if librato.User != "baz" {
		t.Errorf("Incorrect user for librato. Found: '%s', expected: '%s'", librato.User, "baz")
	}

	if librato.Token != "quux" {
		t.Errorf("Incorrect token for librato. Found: '%s', expected: '%s'", librato.Token, "quux")
	}
}

type ClosingHandler struct {
	times, maxCloses int
	data             []byte
	headers          http.Header
}

func (c *ClosingHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c.times++
	if c.times <= c.maxCloses {
		conn, _, _ := w.(http.Hijacker).Hijack()
		conn.Close()
		return
	}

	d, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	c.data = append(c.data, d...)
	c.headers = req.Header
}

func TestLibrato_EOF(t *testing.T) {
	handler := &ClosingHandler{maxCloses: 1}
	server := httptest.NewServer(handler)
	defer server.Close()

	config := GetConfig()
	config.LibratoUrl, _ = url.Parse(server.URL)
	config.NetworkTimeout = 1 * time.Second
	config.LibratoUser = "user"
	config.LibratoToken = "token"

	measurements := make(chan Measurement, 10)
	librato := NewLibratoOutputter(measurements, config)

	if !librato.sendWithBackoff([]byte(`{}`)) {
		t.Errorf("Request should not have errored with a closing handler")
	}

	if handler.times != 2 {
		t.Errorf("Request should have only been tried twice, instead it was tried: %d", handler.times)
	}
}
