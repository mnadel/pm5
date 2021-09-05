package main

import (
	"time"

	log "github.com/sirupsen/logrus"
	"tinygo.org/x/bluetooth"
)

type Client struct {
	config      *Configuration
	device      *PM5Device
	adapter     *bluetooth.Adapter
	subscribers map[byte][]Subscriber
	exitCh      chan struct{}
}

func NewClient(config *Configuration, device *PM5Device) *Client {
	return &Client{
		config:      config,
		device:      device,
		adapter:     bluetooth.DefaultAdapter,
		subscribers: make(map[byte][]Subscriber),
		exitCh:      make(chan struct{}, 1),
	}
}

func (c *Client) Register(char *Characterisic) {
	if val, ok := c.subscribers[char.Message]; ok {
		c.subscribers[char.Message] = append(val, char.Subscriber)
	} else {
		c.subscribers[char.Message] = make([]Subscriber, 1)
		c.subscribers[char.Message][0] = char.Subscriber
	}
}

func (c *Client) Exit() chan struct{} {
	return c.exitCh
}

func (c *Client) Scan() {
	scanResultCh := make(chan bluetooth.ScanResult, 1)

	if c.adapter == nil {
		log.Panic("adapter is nil")
	}

	if err := c.adapter.Enable(); err != nil {
		log.WithError(err).Fatal("cannot enable ble")
	}

	err := c.adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		MetricBLEScans.Add(1)
		MetricLastScan.SetToCurrentTime()

		if result.Address.String() == c.device.DeviceAddress {
			log.WithFields(log.Fields{
				"address": result.Address.String(),
				"name":    result.LocalName(),
				"rssi":    result.RSSI,
			}).Info("found device")

			adapter.StopScan()
			scanResultCh <- result
		}
	})

	if err != nil {
		log.WithError(err).Fatal("error scanning")
	}

	log.Info("starting discovery")
	result := <-scanResultCh

	var device *bluetooth.Device
	device, err = c.adapter.Connect(result.Address, bluetooth.ConnectionParams{})
	if err != nil {
		log.WithError(err).Error("cannot connect")
		c.Exit() <- struct{}{}
	}

	MetricBLEConnects.Add(1)
	log.WithField("address", result.Address.String()).Info("connected")

	srvcs, err := device.DiscoverServices([]bluetooth.UUID{c.device.ServiceUUID})
	if err != nil {
		log.WithError(err).Fatal("cannot discover services")
	}

	if len(srvcs) == 0 {
		log.Fatal("could not find PM5 service")
	}

	characteristicUUIDs := c.device.CharacteristicUUIDs()

	log.WithFields(log.Fields{
		"service": srvcs[0].UUID().String(),
		"chars":   characteristicUUIDs,
	}).Info("found service, looking for chars")

	discoveredCharacteristics, err := srvcs[0].DiscoverCharacteristics(characteristicUUIDs)
	if err != nil {
		log.WithError(err).Fatal("cannot discover characteristics")
	}

	if len(discoveredCharacteristics) != len(characteristicUUIDs) {
		log.WithField("discovered", discoveredCharacteristics).Error("cannot find every characteristic")
	}

	for _, discovered := range discoveredCharacteristics {
		log.WithField("char", discovered.UUID().String())
		c.device.Register(discovered)
	}

	timer := time.NewTimer(c.config.BleReceiveTimeout)
	<-timer.C

	c.Exit() <- struct{}{}
}
