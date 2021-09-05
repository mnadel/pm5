package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"tinygo.org/x/bluetooth"
)

// Characterisic represents a BLE Characteristic and how we process its messages.
type Characterisic struct {
	Name       string
	Message    byte
	UUID       bluetooth.UUID
	Subscriber Subscriber
}

// PM5Device represents the PM5 BLE Peripheral we scan for and listen to.
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

func (c *Characterisic) MessageName() string {
	return fmt.Sprintf("%x", c.Message)
}

// CharacteristicUUIDs gets the list of BLE Characteristic UUIDs we're interested in.
func (d *PM5Device) CharacteristicUUIDs() []bluetooth.UUID {
	arr := make([]bluetooth.UUID, len(d.Characteristics))
	for i, c := range d.Characteristics {
		arr[i] = c.UUID
	}
	return arr
}

// FindCharacteristic returns our Characterisic given a discovered BLE Characterisic UUID.
func (d *PM5Device) FindCharacteristic(uuid bluetooth.UUID) *Characterisic {
	for _, c := range d.Characteristics {
		if c.UUID == uuid {
			return c
		}
	}

	return nil
}

// Register sets up the callbacks for listening to a discovered BLE Characteristic.
func (d *PM5Device) Register(c bluetooth.DeviceCharacteristic) {
	char := d.FindCharacteristic(c.UUID())
	if char == nil {
		log.WithField("uuid", c.UUID()).Error("error looking up characteristic")
		return
	}

	log.WithFields(log.Fields{
		"uuid":    c.UUID().String(),
		"service": char.Name,
		"msg":     char.MessageName(),
	}).Info("subscribing to char's messages")

	c.EnableNotifications(func(buf []byte) {
		MetricMessages.WithLabelValues(char.MessageName()).Add(1)
		char.Subscriber.Notify(buf)
	})
}
