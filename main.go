package main

import (
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
		client.Scan()
	}()

	log.Info("awaiting termination")
	<-client.Exit()

	log.Info("scan terminated")
}
