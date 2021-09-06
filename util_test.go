package main

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestGetPromGaugeValue(t *testing.T) {
	g := prometheus.NewGauge(prometheus.GaugeOpts{})
	g.Set(3.14)
	v := GetPromGaugeValue(&g)

	assert.Equal(t, 3.14, v)
}

func TestGetPromCounterValue(t *testing.T) {
	c := prometheus.NewCounter(prometheus.CounterOpts{})
	c.Add(7)
	v := GetPromCounterValue(&c)

	assert.Equal(t, float64(7), v)
}

func TestContains(t *testing.T) {
	haystack := []string{"a", "b", "c"}
	assert.True(t, Contains(haystack, "a"))
	assert.True(t, Contains(haystack, "b"))
	assert.True(t, Contains(haystack, "c"))
	assert.False(t, Contains(haystack, "d"))
}

func TestHash(t *testing.T) {
	data := []byte{0xc, 0xa, 0xf, 0xe, 0xb, 0xa, 0xb, 0xe}
	assert.Equal(t, "c0821497fd01016258cc47e5d8f8d8dc", Hash(data))
}
