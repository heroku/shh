package mm

import (
	"fmt"
	"os"
	"time"
)

var (
	source = os.Getenv("SHH_SOURCE")
)

type Measurement struct {
	When  time.Time
	What  string
	Value []byte
}

func (m *Measurement) String() string {
	msg := fmt.Sprintf("when=%s measure=%s val=%s", m.When.Format(time.RFC3339Nano), m.What, m.Value)
	if source != "" {
		return fmt.Sprintf("%s source=%s", msg, source)
	}
	return msg
}
