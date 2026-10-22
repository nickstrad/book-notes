// Package clock controls API semantic time only. Process and network timers stay real.
package clock

import (
	"fmt"
	"sync"
	"time"
)

type Clock struct {
	mu    sync.RWMutex
	fixed *time.Time
}

func (c *Clock) Now() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.fixed != nil {
		return *c.fixed
	}
	return time.Now().UTC()
}
func (c *Clock) Change(action string, at time.Time, seconds float64) (time.Time, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now().UTC()
	if c.fixed != nil {
		now = *c.fixed
	}
	switch action {
	case "reset":
		c.fixed = nil
		return now, nil
	case "set":
		if at.IsZero() {
			return now, fmt.Errorf("at required")
		}
		now = at.UTC()
	case "advance":
		d := seconds * float64(time.Second)
		if d >= float64(1<<63-1) || d <= -float64(1<<63-1) {
			return now, fmt.Errorf("advance out of range")
		}
		now = now.Add(time.Duration(d))
	default:
		return now, fmt.Errorf("unknown clock action")
	}
	c.fixed = &now
	return now, nil
}
