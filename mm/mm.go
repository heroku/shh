package mm

import (
	"fmt"
	"github.com/freeformz/shh/utils"
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
	msg := fmt.Sprintf("when=%s measure=%s val=%s", m.Timestamp(), m.Measured(), m.Value)
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

func (m *Measurement) Difference(last string) string {
	currentValue := utils.Atouint64(m.Value)
	lastValue := utils.Atouint64(last)
	// This is a crappy way to handle wraps and resets when we don't know 
	// what the max value is (32, 64 or 128 bit)
	// Leads to a little, loss, but should be minimal overall
	if currentValue < lastValue {
		return m.Value
	}
	return utils.Ui64toa(currentValue - lastValue)
}

func (m *Measurement) Timestamp() string {
	return m.When.Format(time.RFC3339)
}
