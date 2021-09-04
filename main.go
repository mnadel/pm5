package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"tinygo.org/x/bluetooth"
)

var Config *Configuration

func init() {
	Config = NewConfiguration()

	log.SetFormatter(&log.TextFormatter{
		DisableColors:    true,
		FullTimestamp:    true,
		ForceQuote:       true,
		DisableTimestamp: !isTTY(),
	})

	if Config.LogLevel == log.DebugLevel {
		log.SetReportCaller(true)
	}

	log.SetOutput(os.Stdout)
	log.SetLevel(Config.LogLevel)

	log.WithField("config", *Config).Info("loaded configuration")

	StartAdminConsole()
}

func main() {
	pm5device := NewPM5Device()
	adapter := bluetooth.DefaultAdapter
	nullScan := &bluetooth.ScanResult{}

	log.Info("enabling ble stack")
	if err := adapter.Enable(); err != nil {
		log.WithError(err).Fatal("cannot enable ble")
	}

	scanResultCh := make(chan bluetooth.ScanResult, 1)

	log.Info("scanning...")
	err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		MetricBLEScans.Add(1)
		MetricLastScan.SetToCurrentTime()

		if result.Address.String() == pm5device.DeviceAddress {
			log.WithFields(log.Fields{
				"address": result.Address.String(),
				"name":    result.LocalName(),
				"rssi":    result.RSSI,
			}).Info("found device")

			adapter.StopScan()
			scanResultCh <- result
		} else {
			time.Sleep(Config.BleScanFreq)
		}
	})

	if err != nil {
		log.WithError(err).Fatal("error scanning")
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.WithField("signal", sig).Info("signal received from os")
		scanResultCh <- *nullScan
	}()

	log.Info("awaiting device discovery")
	result := <-scanResultCh

	// result was a result of trapping a signal
	if result == *nullScan {
		log.Info("exiting")
		adapter.StopScan()
		os.Exit(0)
	}

	var device *bluetooth.Device
	device, err = adapter.Connect(result.Address, bluetooth.ConnectionParams{})
	if err != nil {
		log.WithError(err).Error("cannot connect")
		return
	}

	MetricBLEConnects.Add(1)
	log.WithField("address", result.Address.String()).Info("connected")

	log.Info("discovering services")
	srvcs, err := device.DiscoverServices([]bluetooth.UUID{pm5device.ServiceUUID})
	if err != nil {
		log.WithError(err).Fatal("cannot discover services")
	}

	if len(srvcs) == 0 {
		log.Fatal("could not find PM5 service")
	}

	srvc := srvcs[0]

	log.WithField("service", srvc.UUID().String()).Info("found service")

	chars, err := srvc.DiscoverCharacteristics([]bluetooth.UUID{pm5device.WorkoutUUID})
	if err != nil {
		log.WithError(err).Fatal("cannot discover characteristics")
	}

	if len(chars) != 1 {
		log.Fatal("cannot find workout characteristic")
	}

	char0039 := chars[0]
	log.WithField("uuid", char0039.UUID().String()).Info("subscribing")

	char0039.EnableNotifications(func(buf []byte) {
		Metric0039Messages.Add(1)
		log.WithField("buf", buf).Info("received data")
	})

	timer := time.NewTimer(Config.BleReceiveTimeout)
	<-timer.C

	log.Info("recv timeout expired")
}
