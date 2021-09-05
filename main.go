package main

import (
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
			"msg":          characteristic.MessageName(),
		}).Info("registering")

		client.Register(characteristic)
	}

	log.Info("starting central")

	client.Listen()

	log.Info("central terminated")
}
