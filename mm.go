package main

import (
	"fmt"
	"strings"
	"time"
)

var (
	MetricNameNormalizer = strings.NewReplacer("#", "_", "-", "_")
)

type MeasurementType int

const (
	CounterType MeasurementType = iota
	GaugeType
	FloatGaugeType
)

type CounterMeasurement struct {
	time   time.Time
	poller string
	what   []string
	value  uint64
	unit UnitType
}

type GaugeMeasurement struct {
	time   time.Time
	poller string
	what   []string
	value  uint64
	unit UnitType
}

type FloatGaugeMeasurement struct {
	time   time.Time
	poller string
	what   []string
	value  float64
	unit UnitType
}

type Measurement interface {
	Name(prefix string) string // Metric name
	Value() interface{}
	StrValue() string // String representation of the value
	Time() time.Time  // the underlying time object.
	Type() MeasurementType
	Unit() string
	UnitAbbr() string
}

func combinedName(prefix, poller string, what []string) string {
	v := fmt.Sprintf("%s.%s", poller, strings.Join(what, "."))
	if prefix != "" {
		v = fmt.Sprintf("%s.%s", prefix, v)
	}
	return MetricNameNormalizer.Replace(v)
}

func (c CounterMeasurement) Name(prefix string) string {
	return combinedName(prefix, c.poller, c.what)
}

func (c CounterMeasurement) StrValue() string {
	return fmt.Sprintf("%d", c.Value)
}

func (c CounterMeasurement) Value() interface{} {
	return c.value
}

func (c CounterMeasurement) Time() time.Time {
	return c.time
}

func (c CounterMeasurement) Type() MeasurementType {
	return CounterType
}

func (c CounterMeasurement) Unit() string {
	return c.unit.Name()
}

func (c CounterMeasurement) UnitAbbr() string {
	return c.unit.Abbr()
}

func (c CounterMeasurement) Difference(l CounterMeasurement) uint64 {
	// This is a crappy way to handle wraps and resets when we don't know
	// what the max value is (32, 64 or 128 bit)
	// Leads to a little, loss, but should be minimal overall
	cv := c.value
	lv := l.value

	if cv < lv {
		return cv
	}
	return cv - lv
}

func (g GaugeMeasurement) Name(prefix string) string {
	return combinedName(prefix, g.poller, g.what)
}

func (g GaugeMeasurement) StrValue() string {
	return fmt.Sprintf("%d", g.value)
}

func (g GaugeMeasurement) Value() interface{} {
	return g.value
}

func (g GaugeMeasurement) Time() time.Time {
	return g.time
}

func (c GaugeMeasurement) Type() MeasurementType {
	return GaugeType
}

func (c GaugeMeasurement) Unit() string {
	return c.unit.Name()
}

func (c GaugeMeasurement) UnitAbbr() string {
	return c.unit.Abbr()
}


func (g FloatGaugeMeasurement) Name(prefix string) string {
	return combinedName(prefix, g.poller, g.what)
}

func (g FloatGaugeMeasurement) StrValue() string {
	return fmt.Sprintf("%f", g.value)
}

func (g FloatGaugeMeasurement) Value() interface{} {
	return g.value
}

func (g FloatGaugeMeasurement) Time() time.Time {
	return g.time
}

func (c FloatGaugeMeasurement) Type() MeasurementType {
	return FloatGaugeType
}

func (c FloatGaugeMeasurement) Unit() string {
	return c.unit.Name()
}

func (c FloatGaugeMeasurement) UnitAbbr() string {
	return c.unit.Abbr()
}

// func (m Measurement) Timestamp() string {
// 	return m.When.Format(time.RFC3339)
// }
