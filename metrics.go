package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	MetricBLEScans = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pm5_ble_scan_total",
		Help: "The number of times we've scanned for a BLE device",
	})

	MetricLastScan = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pm5_ble_scan_last",
		Help: "The time we last scanned for a BLE device",
	})

	MetricBLEConnects = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pm5_ble_connect_total",
		Help: "The number of times we've connected to a BLE device",
	})

	Metric0039Messages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pm5_message_0039_total",
		Help: "The number of times we've received a 0x0039 message",
	})
)
