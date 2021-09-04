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

	go func() {
		log.Info("spawning scanner")

		for {
			client.Scan(Config.BleScanTimeout)
			time.Sleep(Config.BleScanFreq)
		}
	}()

	log.Info("awaiting scanning termination")
	<-client.Exit()

	log.Info("scan terminated")
}
