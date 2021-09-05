package main

import (
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh/terminal"
	"tinygo.org/x/bluetooth"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	log "github.com/sirupsen/logrus"
)

const (
	RFC8601 = "2006-01-02 15:04:05"
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

func mustGetTimezone(name string) *time.Location {
	tz, err := time.LoadLocation(name)
	if err != nil {
		log.WithError(err).WithField("tz", name).Fatal("cannot load timezone")
	}
	return tz
}

func GetPromCounterValue(counter *prometheus.Counter) float64 {
	m := dto.Metric{}
	(*counter).Write(&m)
	return *m.Counter.Value
}

func GetPromGaugeValue(gauge *prometheus.Gauge) float64 {
	c := make(chan prometheus.Metric, 1)
	(*gauge).Collect(c)
	m := dto.Metric{}
	_ = (<-c).Write(&m)

	return *m.Gauge.Value
}
