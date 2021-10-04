package main

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	MAX_LOGFILE_SIZE = 5*2 ^ 20 // 5MB
)

type Configuration struct {
	// LogFile is the path to our logfile
	LogFile string
	// ConfigFile is the path to the files we loaded.
	ConfigFile string
	// LogLevel is the logrus level.
	LogLevel log.Level
	// BleWatchdogDeadline is the max duration between scans we'll tolerate.
	BleWatchdogDeadline time.Duration
	// BleWatchdogDisconnect is the max duration after workout sumary is received before we expect a disconnect.
	BleWatchdogWorkoutDisconnect time.Duration
	// BleWatchdogWorkoutDealine is the max duration after we connect to the PM5 before we expect to receive a workout summary.
	BleWatchdogWorkoutDeadline time.Duration
}

func NewConfiguration() *Configuration {
    viper.SetDefault("log_level", "info")
    viper.SetDefault("log_file", "/var/log/pm5.log")
    viper.SetDefault("ble_watchdog_deadline", "60s")
    viper.SetDefault("ble_watchdog_workout_disconnect", "7m")
    viper.SetDefault("ble_watchdog_workout_deadline", "45m")

	viper.SetConfigName("pm5")
	viper.SetConfigType("yml")
	viper.AddConfigPath("/etc")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("PM5")

	err := viper.ReadInConfig()
	if err != nil {
		log.WithError(err).Fatal("cannot load configuration")
	}

	logLevel, err := log.ParseLevel(viper.GetString("log_level"))
	if err != nil {
		log.WithFields(log.Fields{
			"level": viper.GetString("log_level"),
		}).WithError(err).Fatal("cannot parse log level")
	}

	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
		ForceQuote:    true,
	})

	var logfileMode int

	if info, err := os.Stat(viper.GetString("LOG_FILE")); os.IsNotExist(err) {
		logfileMode = os.O_APPEND
	} else if err != nil {
		log.WithError(err).WithField("file", viper.GetString("LOG_FILE")).Fatal("cannot stat logfile")
	} else if info.Size() >= MAX_LOGFILE_SIZE {
		logfileMode = os.O_TRUNC
	} else {
		logfileMode = os.O_APPEND
	}

	if f, err := os.OpenFile(viper.GetString("LOG_FILE"), os.O_WRONLY|os.O_CREATE|logfileMode, 0644); err != nil {
		panic(err.Error())
	} else {
		log.SetOutput(f)
	}

	log.SetLevel(logLevel)
	log.SetReportCaller(logLevel == log.DebugLevel)

	config := &Configuration{
		BleWatchdogDeadline:          viper.GetDuration("ble_watchdog_deadline"),
		BleWatchdogWorkoutDisconnect: viper.GetDuration("ble_watchdog_workout_disconnect"),
		BleWatchdogWorkoutDeadline:   viper.GetDuration("ble_watchdog_workout_deadline"),
		ConfigFile:                   viper.ConfigFileUsed(),
		LogLevel:                     logLevel,
	}

	cwd, _ := os.Getwd()

	log.WithFields(log.Fields{
		"config": *config,
		"user":   os.Getenv("USER"),
		"cwd":    cwd,
	}).Info("loaded configuration")

	return config
}
