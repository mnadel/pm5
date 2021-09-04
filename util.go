package main

import (
	"os"
	"strconv"

	"golang.org/x/crypto/ssh/terminal"
	"tinygo.org/x/bluetooth"

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
