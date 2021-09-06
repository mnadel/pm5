package main

import (
	log "github.com/sirupsen/logrus"
)

// WorkoutSubscriber receives workout data (0x39 payloads) from the PM5
type WorkoutSubscriber struct {
	config *Configuration
}

func NewWorkoutSubscriber(config *Configuration) *WorkoutSubscriber {
	return &WorkoutSubscriber{
		config: config,
	}
}

func (ws *WorkoutSubscriber) Notify(data []byte) {
	watchdog := NewWatchdog(ws.config)
	watchdog.StartDisconnectMonitor()

	log.Infof("received data: %x", data)

	raw := ReadWorkoutData(data)
	log.WithFields(log.Fields{
		"raw":     raw,
		"message": "workout",
	}).Info("received data")

	decoded := raw.Decode()
	log.WithFields(log.Fields{
		"decoded": decoded,
		"message": "workout",
	}).Info("decoded data")
}
