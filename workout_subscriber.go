package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// WorkoutSubscriber receives workout data (0x39 payloads) from the PM5 Rower.
type WorkoutSubscriber struct {
	config   *Configuration
	database *Database
}

func NewWorkoutSubscriber(config *Configuration) *WorkoutSubscriber {
	return &WorkoutSubscriber{
		config:   config,
		database: NewDatabase(config),
	}
}

func (ws *WorkoutSubscriber) Close() {
	ws.database.Close()
}

func (ws *WorkoutSubscriber) Stats() interface{} {
	return map[string]interface{}{
		"db": ws.database.Stats(),
	}
}

func (ws *WorkoutSubscriber) Notify(data []byte) {
	// i abhor the time-based approach here, but the disconnect callback doesn't seem to get invoked,
	// so a while after this "last" subscriber (argh, till we add more subscribers) we'll force a
	// termination and let systemd restart us.
	// update: tiny-go/bluetooth docs say events can get missed and/or duplicated on linux+bluez, see:
	// https://pkg.go.dev/tinygo.org/x/bluetooth@v0.3.0#Adapter.Scan
	watchdog := NewWatchdog(ws.config)
	watchdog.StartWorkoutDisconnectMonitor()

	log.WithFields(log.Fields{
		"data":    data,
		"message": "workout",
	}).Info("received message")

	if err := ws.database.SaveWorkout(&WorkoutDBRecord{Data: data}); err != nil {
		log.WithError(err).Error("error saving workout in db")
	}

	parsed := ReadWorkoutData(data)
	decoded := parsed.Decode()
	fmt.Println(decoded.AsJSON())
}
