package main

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

const (
	MAX_LOGFILE_SIZE = 5*2 ^ 20 // 5MB
)

type Configuration struct {
	// BleWatchdogDeadline is the max duration between scans we'll tolerate.
	BleWatchdogDeadline time.Duration
	// BleWatchdogDisconnect is the max duration after workout sumary is received before we expect a disconnect.
	BleWatchdogWorkoutDisconnect time.Duration
	// BleWatchdogWorkoutDealine is the max duration after we connect to the PM5 before we expect to receive a workout summary.
	BleWatchdogWorkoutDeadline time.Duration
}

func NewConfiguration() *Configuration {
	logLevel := flag.String("loglevel", "info", "the logrus log level")
	logFile := flag.String("logfile", "/var/log/pm5.log", "path to logfile")
	bleWatchdogDeadline := flag.Duration("scan", time.Second*60, "max duration between scans we'll tolerate")
	bleWatchdogWorkoutDisconnect := flag.Duration("disconn", time.Minute*7, "max duration after workout sumary is received before we expect a disconnect")
	bleWatchdogWorkoutDeadline := flag.Duration("deadline", time.Minute*45, "max duration after we connect to the PM5 before we expect to receive a workout summary")

	flag.Parse()

	parsedLogLevel, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.WithFields(log.Fields{
			"level": logLevel,
		}).WithError(err).Fatal("cannot parse log level")
	}

	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
		ForceQuote:    true,
	})

	var logfileMode int

	if info, err := os.Stat(*logFile); os.IsNotExist(err) {
		logfileMode = os.O_APPEND
	} else if err != nil {
		log.WithError(err).WithField("file", *logFile).Fatal("cannot stat logfile")
	} else if info.Size() >= MAX_LOGFILE_SIZE {
		logfileMode = os.O_TRUNC
	} else {
		logfileMode = os.O_APPEND
	}

	if f, err := os.OpenFile(*logFile, os.O_WRONLY|os.O_CREATE|logfileMode, 0644); err != nil {
		panic(err.Error())
	} else {
		log.SetOutput(f)
	}

	log.SetLevel(parsedLogLevel)
	log.SetReportCaller(parsedLogLevel == log.DebugLevel)

	config := &Configuration{
		BleWatchdogDeadline:          *bleWatchdogDeadline,
		BleWatchdogWorkoutDisconnect: *bleWatchdogWorkoutDisconnect,
		BleWatchdogWorkoutDeadline:   *bleWatchdogWorkoutDeadline,
	}

	cwd, _ := os.Getwd()

	log.WithFields(log.Fields{
		"config": *config,
		"user":   os.Getenv("USER"),
		"cwd":    cwd,
	}).Info("loaded configuration")

	return config
}
