package main

import (
	"tinygo.org/x/bluetooth"
)

type PM5Device struct {
	DeviceAddress string
	ServiceUUID   bluetooth.UUID
	WorkoutUUID   bluetooth.UUID
}

func NewPM5Device(config *Configuration) *PM5Device {
	return &PM5Device{
		DeviceAddress: config.PM5DeviceAddress,
		ServiceUUID:   mustParseUUID("ce060000-43e5-11e4-916c-0800200c9a66"),
		WorkoutUUID:   mustParseUUID("ce060039-43e5-11e4-916c-0800200c9a66"),
	}
}
