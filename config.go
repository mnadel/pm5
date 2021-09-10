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
	BleWatchdogDisconnect time.Duration
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
		DisableTimestamp: !isTTY(),
	})

	log.SetOutput(os.Stdout)
	log.SetLevel(logLevel)
	log.SetReportCaller(logLevel == log.DebugLevel)

	config := &Configuration{
		AdminConsolePort:      viper.GetString("admin_console_port"),
		BleWatchdogDeadline:   viper.GetDuration("ble_watchdog_deadline"),
		BleWatchdogDisconnect: viper.GetDuration("ble_watchdog_disconnect"),
		ConfigFile:            viper.ConfigFileUsed(),
		LogLevel:              logLevel,
	}

	cwd, _ := os.Getwd()

	log.WithFields(log.Fields{
		"config": *config,
		"user":   os.Getenv("USER"),
		"cwd":    cwd,
	}).Info("loaded configuration")

	return config
}
