package main

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

// Watchdog keeps an eye on the system.
type Watchdog struct {
	config *Configuration
}

func NewWatchdog(config *Configuration) *Watchdog {
	return &Watchdog{
		config: config,
	}
}

// StartDisconnectMonitor starts the monitor for a disconnect.
// Returns a channel to cancel the monitor.
func (w *Watchdog) StartDisconnectMonitor() chan<- struct{} {
	cancel := make(chan struct{}, 1)

	go func() {
		deadline := time.Now().Add(w.config.BleWatchdogDisconnect).Format(RFC8601)
		log.WithField("deadline", deadline).Error("starting disconnect watchdog")

		timer := time.NewTimer(w.config.BleWatchdogDisconnect)
		<-timer.C

		select {
		case <-cancel:
			log.Info("canceling disconnect watchdog")
			return
		default:
		}

		log.WithField("elapsed", w.config.BleWatchdogDisconnect).Error("disconnect not received")
		os.Exit(53)
	}()

	return cancel
}

// ScanMonitor monitors BLE scans and terminate self if no progress, let systemd restart us.
// Returns a channel to cancel the monitor.
func (w *Watchdog) ScanMonitor() chan<- struct{} {
	cancel := make(chan struct{}, 1)

	go func() {
	loop:
		for {
			prev := GetPromCounterValue(&MetricBLEScans)
			timer := time.NewTimer(w.config.BleWatchdogDeadline)
			<-timer.C
			current := GetPromCounterValue(&MetricBLEScans)

			select {
			case <-cancel:
				break loop
			default:
			}

			entry := log.WithFields(log.Fields{
				"last_scan": lastScan().Format(RFC8601),
				"prev":      prev,
				"curr":      current,
			})

			if current == prev {
				entry.Fatal("deadlock detected")
				os.Exit(43)
			} else {
				entry.Debug("no deadlock detected")
			}
		}
	}()

	return cancel
}
