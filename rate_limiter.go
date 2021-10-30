package main

import "time"

type RateLimiter struct {
	lastEvent time.Time
	frequency time.Duration
}

func NewRateLimiter(freq time.Duration) *RateLimiter {
	return &RateLimiter{
		frequency: freq,
	}
}

// MaybePerform performs the action if we haven't performed it in the past
// frequency, and returns true if we did perform the action
func (r *RateLimiter) MaybePerform(action func()) bool {
	if time.Since(r.lastEvent) >= r.frequency {
		action()
		r.lastEvent = time.Now()
		return true
	}

	return false
}
