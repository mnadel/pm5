package main

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Configuration struct {
	AdminConsolePort  string
	ConfigFile        string
	BleScanFreq       time.Duration
	LogLevel          log.Level
	BleReceiveTimeout time.Duration
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

	return &Configuration{
		AdminConsolePort:  viper.GetString("admin_console_port"),
		BleScanFreq:       viper.GetDuration("ble_scan_freq"),
		BleReceiveTimeout: viper.GetDuration("ble_recv_timeout"),
		ConfigFile:        viper.ConfigFileUsed(),
		LogLevel:          logLevel,
	}
}
