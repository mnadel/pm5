package main

import (
	log "github.com/sirupsen/logrus"
)

const (
	ERR_BLEDEADLOCK  = 37
	ERR_NODISCONNECT = 41
	ERR_USAGE        = 43
	ERR_CANTSTORE    = 47
	ERR_NOWORKOUT    = 53
)

func main() {
	config := NewConfiguration()
	central := NewCentral(config)

	log.Info("starting central")
	central.Listen()

	log.Info("central exiting")
}
