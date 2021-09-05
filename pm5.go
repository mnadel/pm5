package main

import (
	"tinygo.org/x/bluetooth"
)

type Characterisic struct {
	Name       string
	Message    byte
	UUID       bluetooth.UUID
	Subscriber Subscriber
}

type PM5Device struct {
	DeviceAddress   string
	ServiceUUID     bluetooth.UUID
	Characteristics []*Characterisic
}

func NewPM5Device(config *Configuration) *PM5Device {
	return &PM5Device{
		DeviceAddress: config.PM5DeviceAddress,
		ServiceUUID:   mustParseUUID("ce060000-43e5-11e4-916c-0800200c9a66"),
		Characteristics: []*Characterisic{
			{
				Name:       "workout",
				Message:    0x39,
				UUID:       mustParseUUID("ce060039-43e5-11e4-916c-0800200c9a66"),
				Subscriber: NewWorkoutSubscriber(),
			},
		},
	}
}

func (d *PM5Device) CharacteristicUUIDs() []bluetooth.UUID {
	arr := make([]bluetooth.UUID, len(d.Characteristics))
	for i, c := range d.Characteristics {
		arr[i] = c.UUID
	}
	return arr
}

func (d *PM5Device) FindCharacteristic(uuid bluetooth.UUID) *Characterisic {
	for _, c := range d.Characteristics {
		if c.UUID == uuid {
			return c
		}
	}

	return nil
}
