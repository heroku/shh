package mm

import (
	"fmt"
	"time"
)

type Measurement struct {
	When  time.Time
	What  string
	Value []byte
}

func (m *Measurement) String() string {
	return fmt.Sprintf("when=%s measure=%s val=%s", m.When.Format(time.RFC3339Nano), m.What, m.Value)
}
