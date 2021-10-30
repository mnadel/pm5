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

var (
	CurrentUser = NewUserContext()
)

func main() {
	config := NewConfiguration()
	database := NewDatabase(config)
	central := NewCentral(config, database)
	console := NewAdminServer(config, database, central)
	logbook := NewLogbook(config, database, NewClient())

	go signalHandler(console, central)
	go console.Start()

	syncer := NewSyncer(logbook, database)
	syncer.Start()

	log.Info("starting central")
	central.Listen()

	log.Info("central exiting")

	syncer.Close()
}

func signalHandler(console *AdminServer, central *Central) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	sig := <-c
	log.WithField("signal", sig).Info("term signal received")

	central.Close()
	console.Stop()
}
