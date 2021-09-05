package main

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type Watchdog struct {
	config *Configuration
}

func NewWatchdog(config *Configuration) *Watchdog {
	return &Watchdog{
		config: config,
	}
}

// Monitor for no BLE scans and terminate self if no progress, let systemd restart us.
func (w *Watchdog) Monitor() chan<- struct{} {
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
