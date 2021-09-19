package main

import (
	log "github.com/sirupsen/logrus"
)

// WorkoutSubscriber receives workout data (0x39 payloads) from the PM5 Rower.
type WorkoutSubscriber struct {
	config *Configuration
	dedup  []string
}

func NewWorkoutSubscriber(config *Configuration) *WorkoutSubscriber {
	return &WorkoutSubscriber{
		config: config,
		dedup:  make([]string, 0),
	}
}

func (ws *WorkoutSubscriber) Notify(data []byte) {
	hash := Hash(data)
	log.WithField("hash", hash).Infof("received data: %x", data)

	if Contains(ws.dedup, hash) {
		log.WithField("hash", hash).Info("ignoring duplicate")
		return
	} else {
		ws.dedup = append(ws.dedup, hash)
	}

	// i abhor the time-based approach here, but the disconnect callback doesn't seem to get invoked,
	// so a while after this "last" subscriber (argh, till we add more subscribers) we'll force a
	// termination and let systemd restart us.
	// update: tiny-go/bluetooth docs say events can get missed and/or duplicated on linux+bluez, see:
	// https://pkg.go.dev/tinygo.org/x/bluetooth@v0.3.0#Adapter.Scan
	watchdog := NewWatchdog(ws.config)
	watchdog.StartWorkoutDisconnectMonitor()

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

	logbook, err := NewLogbook(ws.config)
	if err != nil {
		log.WithError(err).Error("cannot create logbook")
		return
	}

	if err := logbook.PublishWorkout(*decoded); err != nil {
		log.WithError(err).WithField("workout", decoded).Error("cannot publish workout")
	}
}
