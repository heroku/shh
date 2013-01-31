package mm

import (
	"fmt"
	"os"
	"strings"
	"time"
)

var (
	source = os.Getenv("SHH_SOURCE")
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
	if source != "" {
		return fmt.Sprintf("%s source=%s", msg, source)
	}
	return msg
}

func (m *Measurement) Measured() string {
	return fmt.Sprintf("%s.%s", m.Poller, strings.Join(m.What, "."))
}

func (m *Measurement) Source() string {
	return source
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
