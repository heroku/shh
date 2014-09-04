package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type HappyHandler struct{}

func (s *HappyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
}

type SleepyHandler struct {
	Amt     time.Duration
	ReqIncr time.Duration
}

func (s *SleepyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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
	handler := &SleepyHandler{2 * time.Second, -400 * time.Millisecond}
	server := httptest.NewServer(handler)
	defer server.Close()

	config := GetConfig()
	config.LibratoUrl = server.URL
	config.LibratoNetworkTimeout = 1 * time.Second
	config.LibratoUser = "user"
	config.LibratoToken = "token"

	measurements := make(chan Measurement, 10)
	librato := NewLibratoOutputter(measurements, config)

	if !librato.sendWithBackoff([]byte(`{}`)) {
		t.Errorf("Request should have completed successfully with a sleepy handler")
	}
}

func TestLibrato_ServerErrorBackoff(t *testing.T) {
	handler := &GrumpyHandler{ResponseCodes: []int{503, 500, 200}}
	server := httptest.NewServer(handler)
	defer server.Close()

	config := GetConfig()
	config.LibratoUrl = server.URL
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
	config.LibratoUrl = server.URL
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
	config.LibratoUrl = server.URL
	config.LibratoUser = "user"
	config.LibratoToken = "token"

	measurements := make(chan Measurement, 10)
	librato := NewLibratoOutputter(measurements, config)

	if librato.sendWithBackoff([]byte(`{}`)) {
		t.Errorf("Retry should not have succeeded due to non-server error.")
	}
}
