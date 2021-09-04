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
		ServiceUUID:   mustParseUUID("CE060000-43E5-11E4-916C-0800200C9A66"),
		WorkoutUUID:   mustParseUUID("CE060039-43E5-11E4-916C-0800200C9A66"),
	}
}
