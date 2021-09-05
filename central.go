package main

import (
	log "github.com/sirupsen/logrus"
	"tinygo.org/x/bluetooth"
)

type Central struct {
	config      *Configuration
	device      *PM5Device
	adapter     *bluetooth.Adapter
	subscribers map[byte][]Subscriber
	exitCh      chan struct{}
}

func NewCentral(config *Configuration) *Central {
	device := NewPM5Device(config)

	central := &Central{
		config:      config,
		device:      device,
		adapter:     bluetooth.DefaultAdapter,
		subscribers: make(map[byte][]Subscriber),
		exitCh:      make(chan struct{}, 1),
	}

	for _, characteristic := range device.Characteristics {
		log.WithFields(log.Fields{
			"service_name": characteristic.Name,
			"msg":          characteristic.MessageName(),
		}).Info("registering")

		central.Register(characteristic)
	}

	return central
}

func (c *Central) Register(char *Characterisic) {
	if val, ok := c.subscribers[char.Message]; ok {
		c.subscribers[char.Message] = append(val, char.Subscriber)
	} else {
		c.subscribers[char.Message] = make([]Subscriber, 1)
		c.subscribers[char.Message][0] = char.Subscriber
	}
}

func (c *Central) Listen() {
	scanResultCh := make(chan bluetooth.ScanResult, 1)

	if c.adapter == nil {
		log.Panic("adapter is nil")
	}

	if err := c.adapter.Enable(); err != nil {
		log.WithError(err).Fatal("cannot enable ble")
	}

	watchdog := NewWatchdog(c.config)
	watchdogCanceler := watchdog.Monitor()

	// this call to Scan won't return until we call adapter.StopScan()
	err := c.adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		MetricBLEScans.Add(1)
		MetricLastScan.SetToCurrentTime()

		if result.Address.String() == c.device.DeviceAddress {
			log.WithFields(log.Fields{
				"address": result.Address.String(),
				"name":    result.LocalName(),
				"rssi":    result.RSSI,
			}).Info("found device")

			scanResultCh <- result
			watchdogCanceler <- struct{}{}
			adapter.StopScan()
		}
	})

	// if we get here, we've either found the BLE Peripheral we're scanning for, or hit an error

	if err != nil {
		log.WithError(err).Fatal("error scanning")
	}

	// retrieve the result from the scan (i.e. the PM5 device)
	result := <-scanResultCh

	// create a channel and callback for syncing on a disconnect event
	disconnectCh := make(chan struct{}, 1)
	c.adapter.SetConnectHandler(func(device bluetooth.Addresser, connected bool) {
		if !connected {
			log.WithField("device", device.String()).Info("detected disconnect")
			disconnectCh <- struct{}{}
		}
	})

	log.Info("connecting to peripheral")

	var device *bluetooth.Device
	device, err = c.adapter.Connect(result.Address, bluetooth.ConnectionParams{})
	if err != nil {
		log.WithError(err).Error("cannot connect")
		return
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

	log.Info("awaiting disconnect")
	<-disconnectCh

	log.Info("central: complete")
}
