package main

import (
	log "github.com/sirupsen/logrus"
)

type Subscriber interface {
	Notify([]byte)
}

type WorkoutSubscriber struct {
}

func NewWorkoutSubscriber() *WorkoutSubscriber {
	return &WorkoutSubscriber{}
}

func (ws *WorkoutSubscriber) Notify(message []byte) {
	log.Infof("received data: %x", message)
}
