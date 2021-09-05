package main

import (
	log "github.com/sirupsen/logrus"
)

func main() {
	config := NewConfiguration()
	startAdminConsole(config)

	device := NewPM5Device(config)
	central := NewCentral(config, device)

	for _, characteristic := range device.Characteristics {
		log.WithFields(log.Fields{
			"service_name": characteristic.Name,
			"msg":          characteristic.MessageName(),
		}).Info("registering")

		central.Register(characteristic)
	}

	log.Info("starting central")

	central.Listen()

	log.Info("central terminated")
}
