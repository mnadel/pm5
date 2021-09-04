package main

import (
	"tinygo.org/x/bluetooth"
)

type PM5Device struct {
	DeviceAddress string
	ServiceUUID   bluetooth.UUID
	WorkoutUUID   bluetooth.UUID
}

func NewPM5Device() *PM5Device {
	return &PM5Device{
		DeviceAddress: "EF:17:C9:1A:D8:18",
		ServiceUUID:   mustParseUUID("ce060000-43e5-11e4-916c-0800200c9a66"),
		WorkoutUUID:   mustParseUUID("ce060039-43e5-11e4-916c-0800200c9a66"),
	}
}
