package main

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Configuration struct {
	// AdminConsolePort is the port to bind the admin webserver to.
	AdminConsolePort string
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
	// LogbookEndpoint is the endpoint of Concept2's Logbook
	LogbookEndpoint string
}

func NewConfiguration() *Configuration {
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
		DisableColors:    true,
		FullTimestamp:    true,
		ForceQuote:       true,
		DisableTimestamp: !IsTTY(),
	})

	log.SetOutput(os.Stdout)
	log.SetLevel(logLevel)
	log.SetReportCaller(logLevel == log.DebugLevel)

	config := &Configuration{
		AdminConsolePort:             viper.GetString("admin_console_port"),
		BleWatchdogDeadline:          viper.GetDuration("ble_watchdog_deadline"),
		BleWatchdogWorkoutDisconnect: viper.GetDuration("ble_watchdog_workout_disconnect"),
		BleWatchdogWorkoutDeadline:   viper.GetDuration("ble_watchdog_workout_deadline"),
		ConfigFile:                   viper.ConfigFileUsed(),
		LogLevel:                     logLevel,
		LogbookEndpoint:              "https://log-dev.concept2.com",
	}

	cwd, _ := os.Getwd()

	log.WithFields(log.Fields{
		"config": *config,
		"user":   os.Getenv("USER"),
		"cwd":    cwd,
	}).Info("loaded configuration")

	return config
}
