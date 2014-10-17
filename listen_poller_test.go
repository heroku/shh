package shh

import (
	"testing"
	"time"
)

func TestListenPollerParser_HappyPath(t *testing.T) {
	var m Measurement
	var err error

	listen := Listen{stats: &ListenStats{counts: make(map[string]interface{})}}

	m, err = listen.parseLine("90210 beverly.hills 10")
	if err != nil {
		t.Errorf("Should have successfully parsed!")
	} else {
		if c, ok := m.(CounterMeasurement); !ok {
			t.Errorf("Should have returned a CounterMeasurement, was= %u", m)
		} else if c.value != 10 {
			t.Errorf("Value should equal 10")
		} else if c.time != time.Unix(90210, 0) {
			t.Errorf("Time should be equal to Unix(90210), was=%s", c.time)
		} else if c.unit.Name() != "" {
			t.Errorf("Unit should be empty, was=%u", c.unit)
		} else if len(c.what) != 1 {
			t.Errorf("Metric name should have 1 component, was=%d", len(c.what))
		}
	}

	m, err = listen.parseLine("2014-10-13T22:00:16Z beverly.hills 10")
	if err != nil {
		t.Errorf("Should have successfully parsed!: got=%s", err)
	} else {
		if c, ok := m.(CounterMeasurement); !ok {
			t.Errorf("Should have returned a CounterMeasurement, was= %u", m)
			t.Errorf("Value should equal 10")
		} else if c.time != time.Date(2014, 10, 13, 22, 0, 16, 0, time.UTC) {
			t.Errorf("Time should be equal to 2014-10-13T22:00:16Z, was=%s", c.time)
		}
	}

	m, err = listen.parseLine("90210 beverly.hills 10 g")
	if err != nil {
		t.Errorf("Should have successfully parsed!")
	} else {
		if c, ok := m.(GaugeMeasurement); !ok {
			t.Errorf("Should have returned a GaugeMeasurement, was= %u", c)
		}
	}

	m, err = listen.parseLine("90210 beverly.hills 10 c")
	if err != nil {
		t.Errorf("Should have successfully parsed!")
	} else {
		if c, ok := m.(CounterMeasurement); !ok {
			t.Errorf("Should have returned a CounterMeasurement, was=%u", c)
		}
	}

	m, err = listen.parseLine("90210 beverly.hills 10 c Millionaires")
	if err != nil {
		t.Errorf("Should have successfully parsed!")
	} else {
		if m.Unit().Name() != "Millionaires" || m.Unit().Abbr() != "" {
			t.Errorf("Unit should have been name=Millionaires, abbr=, was=%u", m)
		}
	}

	m, err = listen.parseLine("90210 beverly.hills 10 c Millionaires,$$")
	if err != nil {
		t.Errorf("Should have successfully parsed!")
	} else {
		if m.Unit().Name() != "Millionaires" || m.Unit().Abbr() != "$$" {
			t.Errorf("Unit should have been name=Millionaires, abbr=$$, was=%u", m)
		}
	}
}

func TestListenPollerParser_Errors(t *testing.T) {
	var err error

	listen := Listen{stats: &ListenStats{counts: make(map[string]interface{})}}

	failure_cases := []string{
		"timestamp metric",
		"timestamp metric value",
		"2014-10-13 22:00:16 non.compliant.ts 10",
		"2014-10-13T22:00:16Z - 10",
		"2014-10-13T22:00:16Z 10",
		"2014-10-13T22:00:16Z negative.counter -1020 c",
		"2014-10-13T22:00:16Z bad.type 10 q",
		"2014-10-13T22:00:16Z malformed.type 10g",
		"2014-10-13T22:00:16Z bad.unit 10 c Bad Unit",
		"2014-10-13T22:00:16Z bad.abbr 10 c BadAbbr,88888",
		"2014-10-13T22:00:16Z malformed.unit 10 c Malform:m",
		"2014-10-13T22:00:16Z malformed.meta 10 c:Malform,m",
	}

	for _, fail := range failure_cases {
		if _, err = listen.parseLine(fail); err == nil {
			t.Errorf("%q should have failed, but passed instead!", fail)
		}
	}
}
