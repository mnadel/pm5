package main

import (
	"time"

	log "github.com/sirupsen/logrus"
)

var Config *Configuration

func init() {
	Config = NewConfiguration()
	StartAdminConsole()
}

func main() {
	client := NewClient()
	exitCh := make(chan struct{}, 1)

	go func() {
		log.Info("spawning scanner")

		for {
			client.Scan(Config.BleScanTimeout, exitCh)
			time.Sleep(Config.BleScanFreq)
		}
	}()

	log.Info("awaiting scanning termination")
	<-exitCh

	log.Info("scan terminated")
}
