package main

import (
	"os"
	"strconv"

	"golang.org/x/crypto/ssh/terminal"

	log "github.com/sirupsen/logrus"
)

const (
	RFC8601 = "2006-01-02 15:04:05"
)

// ShouldParseAtoi parses str into an int, and returns 0 if there's an error
func ShouldParseAtoi(str string) int {
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

func must(action string, err error) {
	if err != nil {
		log.WithError(err).Fatal(action)
	}
}
