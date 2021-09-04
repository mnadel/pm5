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
	log.Debug("creating client")
	client := NewClient()

	log.Debug("looping...")

	for {
		log.Debug("scanning...")
		client.Scan(Config.BleScanTimeout)

		log.Debug("sleeping before next scan")
		time.Sleep(Config.BleScanFreq)
	}
}
