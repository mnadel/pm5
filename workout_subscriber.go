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
	// abhor the time-based approach here, but the disconnect callback doesn't
	// seem to get invoked, so while after this "last" subscriber (argh, till we add more subscribers)
	// we'll force a termination and let systemd restart us
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
