package main

import (
	"fmt"
	"strings"
	"time"
)

var (
	MetricNameNormalizer = strings.NewReplacer("#", "_", "-", "_")
)

type Measurement struct {
	When   time.Time
	Poller string
	What   []string
	Value  interface{}
}

func (m *Measurement) SValue() string {
	switch m.Value.(type) {
	case float64:
		return fmt.Sprintf("%f", m.Value.(float64))
	case uint64:
		return fmt.Sprintf("%d", m.Value.(uint64))
	}
	return ""
}

func (m *Measurement) Measured(prefix string) string {
	v := fmt.Sprintf("%s.%s", m.Poller, strings.Join(m.What, "."))
	if prefix != "" {
		v = fmt.Sprintf("%s.%s", prefix, v)
	}
	return MetricNameNormalizer.Replace(v)
}

func (current *Measurement) Difference(last *Measurement) uint64 {
	// This is a crappy way to handle wraps and resets when we don't know
	// what the max value is (32, 64 or 128 bit)
	// Leads to a little, loss, but should be minimal overall
	cv := current.Value.(uint64)
	lv := last.Value.(uint64)
	if cv < lv {
		return cv
	}
	return cv - lv
}

func (m *Measurement) Timestamp() string {
	return m.When.Format(time.RFC3339)
}

func (m *Measurement) UnixNano() int64 {
	return m.When.UnixNano()
}

func (m *Measurement) Unix() int64 {
	return m.When.Unix()
}
