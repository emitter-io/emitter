package utils

import (
	"time"
)

// Repeat creates a ticker which tickes periodically
func Repeat(action func(), interval time.Duration, closing chan bool) *time.Ticker {
	action()
	timer := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-closing:
				timer.Stop()
				return
			case <-timer.C:
				action()
			}
		}
	}()
	return timer
}
