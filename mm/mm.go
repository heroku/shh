package mm

import (
	"fmt"
	"github.com/freeformz/shh/config"
	"strings"
	"time"
)

type Measurement struct {
	When   time.Time
	Poller string
	What   []string
	Value  interface{}
}

func (m *Measurement) String() string {
	msg := fmt.Sprintf("when=%s measure=%s", m.Timestamp(), m.Measured())
	switch m.Value.(type) {
	case float64:
		msg = fmt.Sprintf("%s val=%f", msg, m.Value.(float64))
	case uint64:
		msg = fmt.Sprintf("%s val=%d", msg, m.Value.(uint64))
	}
	if config.Source != "" {
		return fmt.Sprintf("%s source=%s", msg, config.Source)
	}
	return msg
}

func (m *Measurement) Measured() string {
	v := fmt.Sprintf("%s.%s", m.Poller, strings.Join(m.What, "."))
	if config.Prefix != "" {
		v = fmt.Sprintf("%s.%s", config.Prefix, v)
	}
	return v
}

func (m *Measurement) Source() string {
	return config.Source
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
