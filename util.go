package main

import (
	"crypto/md5"
	"encoding/binary"
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

// ShouldParseAtoi attempts to parse str to an int, otherwise returns 0.
func ShouldParseAtoi(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		log.WithError(err).WithField("str", str).Error("cannot parse")
		return 0
	}
	return i
}

// IsTTY returns true if the program is running from a terminal.
func IsTTY() bool {
	return terminal.IsTerminal(int(os.Stdin.Fd()))
}

// MustParseUUID parses a string UUID into a bluetooth.UUID, or panics.
func MustParseUUID(uuid string) bluetooth.UUID {
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

// Contains returns true if needle is an element in haystack.
func Contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

// Hash returns a string representation of a hash of the data.
func Hash(data []byte) string {
	return fmt.Sprintf("%x", md5.Sum(data))
}

// Itob encodes a uint64 into a byte array
func Itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

// MustInt takes an (int, error) and either panics or returns the int
func MustInt(val int, err error) int {
	if err != nil {
		panic(err.Error())
	}

	return val
}

// Panic panics with a formatted message
func Panic(err error, msg string, args ...interface{}) {
	panic(fmt.Sprintf(msg+": %v", append(args, err)))
}

// MustGetCwd gets the cwd w/o returning an error (but could be "" if there was one)
func MustGetCwd() string {
	cwd, _ := os.Getwd()
	return cwd
}
