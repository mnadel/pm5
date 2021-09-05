package main

import (
	"time"

	log "github.com/sirupsen/logrus"
	"tinygo.org/x/bluetooth"
)

type Client struct {
	config  *Configuration
	device  *PM5Device
	adapter *bluetooth.Adapter
	exitCh  chan struct{}
}

func NewClient(config *Configuration, device *PM5Device) *Client {
	return &Client{
		config:  config,
		device:  device,
		adapter: bluetooth.DefaultAdapter,
		exitCh:  make(chan struct{}, 1),
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
	log.WithField("address", result.Address.String()).Debug("connected")

	log.Debug("discovering services")
	srvcs, err := device.DiscoverServices([]bluetooth.UUID{c.device.ServiceUUID})
	if err != nil {
		log.WithError(err).Fatal("cannot discover services")
	}

	if len(srvcs) == 0 {
		log.Fatal("could not find PM5 service")
	}

	srvc := srvcs[0]

	log.WithField("service", srvc.UUID().String()).Debug("found service")

	chars, err := srvc.DiscoverCharacteristics([]bluetooth.UUID{c.device.WorkoutUUID})
	if err != nil {
		log.WithError(err).Fatal("cannot discover characteristics")
	}

	if len(chars) != 1 {
		log.Fatal("cannot find workout characteristic")
	}

	char0039 := chars[0]
	log.WithField("uuid", char0039.UUID().String()).Debug("subscribing")

	char0039.EnableNotifications(func(buf []byte) {
		Metric0039Messages.Add(1)
		log.Infof("received data: %x", buf)
	})

	timer := time.NewTimer(c.config.BleReceiveTimeout)
	<-timer.C

	c.Exit() <- struct{}{}
}
