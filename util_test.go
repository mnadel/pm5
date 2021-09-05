package main

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestMustGetTimezone(t *testing.T) {
	assert.NotNil(t, mustGetTimezone("America/Chicago"))
}

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
