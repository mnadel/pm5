package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLmits(t *testing.T) {
	var count int

	incrementer := func() {
		count++
	}

	limiter := NewRateLimiter(time.Second * 1)

	// first call will succeed
	assert.True(t, limiter.MaybePerform(incrementer))
	assert.Equal(t, 1, count)

	// next call won't
	assert.False(t, limiter.MaybePerform(incrementer))
	assert.Equal(t, 1, count)

	// but if we wait long enough, it'll succeed again
	time.Sleep(time.Second * 2)
	assert.True(t, limiter.MaybePerform(incrementer))
	assert.Equal(t, 2, count)
}
