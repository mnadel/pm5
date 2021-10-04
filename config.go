package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

const (
	MAX_LOGFILE_SIZE = 5*2 ^ 20 // 5MB
)

type Configuration struct {
	// DBFile is the path to our database file
	DBFile string
	// BleWatchdogDeadline is the max duration between scans we'll tolerate.
	BleWatchdogDeadline time.Duration
	// BleWatchdogDisconnect is the max duration after workout sumary is received before we expect a disconnect.
	BleWatchdogWorkoutDisconnect time.Duration
	// BleWatchdogWorkoutDealine is the max duration after we connect to the PM5 before we expect to receive a workout summary.
	BleWatchdogWorkoutDeadline time.Duration
}

func NewTestConfiguration() *Configuration {
	f, err := ioutil.TempFile(os.TempDir(), "gotest-")
	if err != nil {
		panic(err.Error())
	}

	return &Configuration{
		DBFile: f.Name(),
	}
}

func NewConfiguration() *Configuration {
	initialize := flag.Bool("init", false, "initialize the given config (make directories, etc)")
	logLevel := flag.String("loglevel", "info", "the logrus log level")
	logFile := flag.String("logfile", "/var/log/pm5.log", "path to logfile")
	dbFile := flag.String("dbfile", "/var/run/pm5/pm5.boltdb", "path to db file")
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

	if *logFile == "-" {
		log.SetOutput(os.Stdout)
	} else {
		fsm := &FileSizeManager{}
		if f, err := fsm.OpenFile(*logFile, MAX_LOGFILE_SIZE); err != nil {
			panic(err.Error())
		} else {
			log.SetOutput(f)
		}
	}

	log.SetLevel(parsedLogLevel)
	log.SetReportCaller(parsedLogLevel == log.DebugLevel)

	if *initialize {
		logFileDirectory := filepath.Dir(*logFile)
		log.WithField("dir", logFileDirectory).Info("ensuring directory")
		os.MkdirAll(logFileDirectory, 0755)

		dbFileDirectory := filepath.Dir(*dbFile)
		log.WithField("dir", dbFileDirectory).Info("ensuring directory")
		os.MkdirAll(dbFileDirectory, 0755)

		os.Exit(0)
	}

	config := &Configuration{
		DBFile:                       *dbFile,
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
