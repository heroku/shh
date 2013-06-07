package utils

import "time"

func PreTick(d time.Duration) <-chan time.Time {
	c := make(chan time.Time, 1)
	c <- time.Now()

	go func() {
		ticks := time.Tick(d)
		for tick := range ticks {
			c <- tick
		}
	}()

	return c
}
