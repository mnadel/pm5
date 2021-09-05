package main

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Configuration struct {
	AdminConsolePort    string
	ConfigFile          string
	PM5DeviceAddress    string
	LogLevel            log.Level
	BleReceiveTimeout   time.Duration
	BleWatchdogDeadline time.Duration
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
		AdminConsolePort:    viper.GetString("admin_console_port"),
		PM5DeviceAddress:    viper.GetString("pm5_device_addr"),
		ConfigFile:          viper.ConfigFileUsed(),
		LogLevel:            logLevel,
		BleReceiveTimeout:   viper.GetDuration("ble_recv_timeout"),
		BleWatchdogDeadline: viper.GetDuration("ble_watchdog_deadline"),
	}

	cwd, _ := os.Getwd()

	log.WithFields(log.Fields{
		"config": *config,
		"user":   os.Getenv("USER"),
		"cwd":    cwd,
	}).Info("loaded configuration")

	return config
}
