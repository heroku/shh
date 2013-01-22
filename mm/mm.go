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

type MeasurementType int

const (
	GAUGE   = iota // ex. speedometer reading
	COUNTER = iota // ex. i/o operations completed
)

type Measurement struct {
	When   time.Time
	Poller string
	What   []string
	Value  string
	Type   MeasurementType
}

func (m *Measurement) String() string {
	msg := fmt.Sprintf("when=%s measure=%s val=%s", m.When.Format(time.RFC3339), m.Measured(), m.Value)
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
