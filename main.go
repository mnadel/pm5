package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

const (
	ERR_BLEDEADLOCK  = 37
	ERR_NODISCONNECT = 41
	ERR_CANTSTORE    = 47
	ERR_NOWORKOUT    = 53
)

func main() {
	config := NewConfiguration()
	database := NewDatabase(config)
	central := NewCentral(config, database)
	logbook := NewLogbook(config, database, NewClient())

	go signalHandler(central)
	go AdminConsole(config, central)

	syncer := NewSyncer(logbook, database)
	syncer.Start()

	log.Info("starting central")
	central.Listen()

	log.Info("central exiting")

	syncer.Close()
}

func signalHandler(central *Central) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	sig := <-c
	log.WithField("signal", sig).Info("term signal received")

	central.Close()
}
