package main

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"tinygo.org/x/bluetooth"
)

// Characterisic represents a BLE Characteristic we subscribe to receiving.
type Characterisic struct {
	Name       string
	Message    byte
	UUID       bluetooth.UUID
	Subscriber Subscriber
}

// PM5Device represents the PM5 BLE Peripheral we scan for and listen to.
type PM5Device struct {
	DeviceAddress         string
	RowingServiceUUID     bluetooth.UUID
	RowingCharacteristics []*Characterisic
}

func NewPM5Device(config *Configuration, database *Database) *PM5Device {
	return &PM5Device{
		RowingServiceUUID: MustParseUUID("ce060030-43e5-11e4-916c-0800200c9a66"),
		RowingCharacteristics: []*Characterisic{
			{
				Name:       "workout",
				Message:    0x39,
				UUID:       MustParseUUID("ce060039-43e5-11e4-916c-0800200c9a66"),
				Subscriber: NewWorkoutSubscriber(config, database),
			},
		},
	}
}

// MessageName returns a text representation of the byte offset of the Characteristic's message.
func (c *Characterisic) MessageName() string {
	return fmt.Sprintf("%x", c.Message)
}

func (d *PM5Device) Stats() map[string]interface{} {
	s := make(map[string]interface{})

	for _, rc := range d.RowingCharacteristics {
		s[rc.Name] = rc.Subscriber.Stats()
	}

	return s
}

func (d *PM5Device) Close() {
	for _, v := range d.RowingCharacteristics {
		log.WithField("sub", v.Name).Info("closing subscriber")
		v.Subscriber.Close()
	}
}

// IsPM5 returns true if the parameter represents a PM5 rowing machine.
func (d *PM5Device) IsPM5(r bluetooth.ScanResult) bool {
	return IsPM5(r.LocalName())
}

// RowingCharacteristicUUIDs gets the list of rowing-specific BLE Characteristic UUIDs we're interested in.
func (d *PM5Device) RowingCharacteristicUUIDs() []bluetooth.UUID {
	arr := make([]bluetooth.UUID, len(d.RowingCharacteristics))
	for i, c := range d.RowingCharacteristics {
		arr[i] = c.UUID
	}
	return arr
}

// FindRowingCharacteristic returns our rowing Characterisic given a discovered BLE Characterisic UUID.
func (d *PM5Device) FindRowingCharacteristic(uuid bluetooth.UUID) *Characterisic {
	for _, c := range d.RowingCharacteristics {
		if c.UUID == uuid {
			return c
		}
	}

	return nil
}

// Register sets up the callbacks for listening to a discovered BLE Characteristic.
func (d *PM5Device) Register(c bluetooth.DeviceCharacteristic) {
	char := d.FindRowingCharacteristic(c.UUID())
	if char == nil {
		log.WithField("uuid", c.UUID()).Error("error looking up characteristic")
		return
	}

	log.WithFields(log.Fields{
		"uuid":           c.UUID().String(),
		"service":        char.Name,
		"characteristic": char.MessageName(),
	}).Info("subscribing to messages")

	c.EnableNotifications(func(buf []byte) {
		MetricMessages.WithLabelValues(char.MessageName()).Add(1)
		char.Subscriber.Notify(buf)
	})
}

// IsPM5 returns true if the BLE local name represents a PM5 rower device.
func IsPM5(localName string) bool {
	return strings.Contains(localName, "PM5") && strings.Contains(localName, "Row")
}
