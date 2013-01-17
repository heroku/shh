package mm

import (
	"fmt"
	"os"
	"strings"
)

var (
	source = os.Getenv("SHH_SOURCE")
)

type Measurement struct {
	Poller string
	What   []string
	Value  string
}

func (m *Measurement) String() string {
	msg := fmt.Sprintf("measure=%s.%s val=%s", m.Poller, strings.Join(m.What, "."), m.Value)
	if source != "" {
		return fmt.Sprintf("%s source=%s", msg, source)
	}
	return msg
}
