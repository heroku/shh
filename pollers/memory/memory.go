package memory

import (
	"shh/mm"
	"time"
)

const (
	Name string = "memory"
)

func Poll(now time.Time, measurements chan *mm.Measurement) {
	measurements <- &mm.Measurement{now, "memory.used", []byte("12")}
}
