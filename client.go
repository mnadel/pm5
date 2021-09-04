package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"tinygo.org/x/bluetooth"
)

type Client struct {
	device  *PM5Device
	adapter *bluetooth.Adapter
}

func NewClient() *Client {
	return &Client{
		device:  NewPM5Device(),
		adapter: bluetooth.DefaultAdapter,
	}
}

func (c *Client) Scan(timeout time.Duration, exitCh chan struct{}) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	doneChan := make(chan struct{}, 1)
	scanResultCh := make(chan bluetooth.ScanResult, 1)

	err := c.adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		if t, _ := ctx.Deadline(); time.Now().After(t) {
			log.Debug("surpassed timeout")
			cancel()
			return
		}

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

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.WithField("signal", sig).Debug("signal received from os")
		doneChan <- struct{}{}
	}()

	var result bluetooth.ScanResult

	log.Debug("awaiting discovery")

	select {
	case <-doneChan:
		c.adapter.StopScan()
		cancel()
		exitCh <- struct{}{}
	case result = <-scanResultCh:
		break
	}

	var device *bluetooth.Device
	device, err = c.adapter.Connect(result.Address, bluetooth.ConnectionParams{})
	if err != nil {
		log.WithError(err).Error("cannot connect")
		exitCh <- struct{}{}
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
		log.WithField("buf", buf).Info("received data")
	})

	timer := time.NewTimer(Config.BleReceiveTimeout)
	<-timer.C

	// wait for timer or signal
	select {
	case <-doneChan:
		break
	case <-timer.C:
		log.Debug("recv timeout expired")
	}

	c.adapter.StopScan()
	cancel()

	exitCh <- struct{}{}
}
