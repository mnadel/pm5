package main

import (
	log "github.com/sirupsen/logrus"
)

func main() {
	config := NewConfiguration()
	startAdminConsole(config)

	central := NewCentral(config)

	log.Info("starting central")
	central.Listen()

	log.Info("central terminated")
}
