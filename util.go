package main

import (
	"crypto/md5"
	"fmt"
	"os"
	"strconv"

	"golang.org/x/crypto/ssh/terminal"
	"tinygo.org/x/bluetooth"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	log "github.com/sirupsen/logrus"
)

const (
	ISO8601 = "2006-01-02 15:04:05"
)

func shouldParseAtoi(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		log.WithError(err).WithField("str", str).Error("cannot parse")
		return 0
	}
	return i
}

func isTTY() bool {
	return terminal.IsTerminal(int(os.Stdin.Fd()))
}

func mustParseUUID(uuid string) bluetooth.UUID {
	u, err := bluetooth.ParseUUID(uuid)
	if err != nil {
		log.WithError(err).WithField("uuid", uuid).Fatal("cannot parse uuid")
	}
	return u
}

// GetPromCounterValue gets the value of a counter.
func GetPromCounterValue(counter *prometheus.Counter) float64 {
	m := dto.Metric{}
	(*counter).Write(&m)
	return *m.Counter.Value
}

// GetPromGaugeValue gets the value of a gauge.
func GetPromGaugeValue(gauge *prometheus.Gauge) float64 {
	c := make(chan prometheus.Metric, 1)
	(*gauge).Collect(c)
	m := dto.Metric{}
	_ = (<-c).Write(&m)

	return *m.Gauge.Value
}

func Contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func Hash(data []byte) string {
	return fmt.Sprintf("%x", md5.Sum(data))
}
