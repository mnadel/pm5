package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

func main() {
	config := NewConfiguration()
	startAdminConsole(config)

	device := NewPM5Device(config)
	client := NewClient(config, device)

	for _, characteristic := range device.Characteristics {
		log.WithFields(log.Fields{
			"service_name": characteristic.Name,
			"msg":          fmt.Sprintf("%x", characteristic.Message),
		}).Info("registering")

		client.Register(characteristic)
	}

	go func() {
		log.Info("spawning scanner")
		client.Scan()
	}()

	log.Info("awaiting termination")
	<-client.Exit()

	log.Info("scan terminated")
}
