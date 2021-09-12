package main

import (
	log "github.com/sirupsen/logrus"
	"tinygo.org/x/bluetooth"
)

// Central represents our BLE Central device.
type Central struct {
	config      *Configuration
	device      *PM5Device
	adapter     *bluetooth.Adapter
	subscribers map[byte][]Subscriber
}

func NewCentral(config *Configuration) *Central {
	central := &Central{
		config:      config,
		device:      NewPM5Device(config),
		adapter:     bluetooth.DefaultAdapter,
		subscribers: make(map[byte][]Subscriber),
	}

	for _, characteristic := range central.device.RowingCharacteristics {
		log.WithFields(log.Fields{
			"service_name": characteristic.Name,
			"char_id":      characteristic.MessageName(),
		}).Info("registering")

		central.Register(characteristic)
	}

	return central
}

// Register registers a Characteristic to which our Central will receive messages.
func (c *Central) Register(char *Characterisic) {
	if val, ok := c.subscribers[char.Message]; ok {
		c.subscribers[char.Message] = append(val, char.Subscriber)
	} else {
		c.subscribers[char.Message] = make([]Subscriber, 1)
		c.subscribers[char.Message][0] = char.Subscriber
	}
}

// Listen is a blocking call that scans for and connects to a PM5 Rower BLE Peripheral, and then awaits
// Characteristic messages to be received and processes them.
func (c *Central) Listen() {
	scanResultCh := make(chan bluetooth.ScanResult, 1)

	if c.adapter == nil {
		log.Panic("adapter is nil")
	}

	if err := c.adapter.Enable(); err != nil {
		log.WithError(err).Fatal("cannot enable ble")
	}

	watchdog := NewWatchdog(c.config)
	watchdogCanceler := watchdog.ScanMonitor()

	// this call to Scan won't return until we call adapter.StopScan()
	err := c.adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		MetricBLEScans.Add(1)
		MetricLastScan.SetToCurrentTime()

		if c.device.IsPM5(result) {
			log.WithFields(log.Fields{
				"address": result.Address.String(),
				"name":    result.LocalName(),
				"rssi":    result.RSSI,
			}).Info("found pm5")

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
	// this doesn't seem to work (it's never called) on my RPi 3B+
	// see Watchdog for the workaround
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

	srvcs, err := device.DiscoverServices([]bluetooth.UUID{c.device.RowingServiceUUID})
	if err != nil {
		log.WithError(err).Fatal("cannot discover rowing service")
	}

	if len(srvcs) == 0 {
		log.Fatal("could not find PM5 rowing service")
	}

	characteristicUUIDs := c.device.RowingCharacteristicUUIDs()

	log.WithFields(log.Fields{
		"service": srvcs[0].UUID().String(),
		"chars":   characteristicUUIDs,
	}).Info("found rowing service, looking for rowing characteristics")

	discoveredCharacteristics, err := srvcs[0].DiscoverCharacteristics(characteristicUUIDs)
	if err != nil {
		log.WithError(err).Fatal("cannot discover rowing characteristics")
	}

	if len(discoveredCharacteristics) != len(characteristicUUIDs) {
		log.WithField("discovered", discoveredCharacteristics).Error("found subset of characteristics")
	}

	for _, discovered := range discoveredCharacteristics {
		log.WithField("characteristic", discovered.UUID().String())
		c.device.Register(discovered)
	}

	log.Info("awaiting disconnect")
	<-disconnectCh

	log.Info("central: complete")
}
