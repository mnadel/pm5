package main

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Configuration struct {
	AdminConsolePort  string
	ConfigFile        string
	LogLevel          log.Level
	BleScanFreq       time.Duration
	BleReceiveTimeout time.Duration
	BleScanTimeout    time.Duration
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

	if logLevel == log.DebugLevel {
		log.SetReportCaller(true)
	}

	log.SetOutput(os.Stdout)
	log.SetLevel(logLevel)

	config := &Configuration{
		AdminConsolePort:  viper.GetString("admin_console_port"),
		BleReceiveTimeout: viper.GetDuration("ble_recv_timeout"),
		ConfigFile:        viper.ConfigFileUsed(),
		LogLevel:          logLevel,
	}

	log.WithField("config", *config).Info("loaded configuration")

	return config
}
