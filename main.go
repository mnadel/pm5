package main

import (
	log "github.com/sirupsen/logrus"
)

const (
	ERR_BLEDEADLOCK  = 37
	ERR_NODISCONNECT = 41
)

func main() {
	config := NewConfiguration()
	startAdminConsole(config)

	central := NewCentral(config)

	log.Info("starting central")
	central.Listen()

	log.Info("central exiting")
}
