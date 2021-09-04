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

	for {
		client.Scan(Config.BleScanTimeout)

		log.Debug("sleeping before next scan")
		time.Sleep(Config.BleScanFreq)
	}
}
