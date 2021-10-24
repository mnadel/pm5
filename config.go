package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

const (
	MAX_LOGFILE_SIZE   = 5*2 ^ 20 // 5MB
	PM5_OAUTH_APPID    = "ymMRExBCsS6HqDm9ShMEPRvpR3Hh2DPb3FTtiazX"
	PM5_OAUTH_CALLBACK = "https://auth.pm5-book.workers.dev/c2"
)

type Configuration struct {
	// LogbookHost is the DNS host of the Logbook service
	LogbookHost string
	// DBFile is the path to our database file
	DBFile string
	// BleWatchdogDeadline is the max duration between scans we'll tolerate.
	BleWatchdogDeadline time.Duration
	// BleWatchdogDisconnect is the max duration after workout sumary is received before we expect a disconnect.
	BleWatchdogWorkoutDisconnect time.Duration
	// BleWatchdogWorkoutDealine is the max duration after we connect to the PM5 before we expect to receive a workout summary.
	BleWatchdogWorkoutDeadline time.Duration
	// AdminConsolePort is the port to which we're attaching our web console
	AdminConsolePort string
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
	printDB := flag.Bool("dbdump", false, "print the contents of the database")
	refresh := flag.Bool("refresh", false, "get a new refresh token")
	authURL := flag.Bool("authurl", false, "print the auth url")

	host := flag.String("host", "log.concept2.com", "specify the logbook service hostname")
	auth := flag.String("auth", "", "set the auth token in the form of id:secret")
	dbFile := flag.String("dbfile", "/var/run/pm5/pm5.boltdb", "path to db file")
	logFile := flag.String("logfile", "/var/log/pm5.log", "path to logfile")
	logLevel := flag.String("loglevel", "info", "the logrus log level")
	port := flag.String("port", ":2112", "web console port")

	bleWatchdogWorkoutDeadline := flag.Duration("deadline", time.Minute*35, "max duration after we connect to the PM5 before we expect to receive a workout summary")
	bleWatchdogWorkoutDisconnect := flag.Duration("disconn", time.Minute*7, "max duration after workout sumary is received before we expect a disconnect")
	bleWatchdogDeadline := flag.Duration("scan", time.Second*60, "max duration between scans we'll tolerate")

	flag.Parse()

	if *authURL {
		uriFmt := "https://%s/oauth/authorize?client_id=%s&scope=results:write&response_type=code&redirect_uri=%s"
		uri := fmt.Sprintf(uriFmt, *host, PM5_OAUTH_APPID, PM5_OAUTH_CALLBACK)
		fmt.Println("please visit the below url")
		fmt.Println(uri)

		os.Exit(0)
	}

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
		LogbookHost:                  *host,
		DBFile:                       *dbFile,
		BleWatchdogDeadline:          *bleWatchdogDeadline,
		BleWatchdogWorkoutDisconnect: *bleWatchdogWorkoutDisconnect,
		BleWatchdogWorkoutDeadline:   *bleWatchdogWorkoutDeadline,
		AdminConsolePort:             *port,
	}

	cwd, _ := os.Getwd()

	log.WithFields(log.Fields{
		"config": *config,
		"user":   os.Getenv("USER"),
		"cwd":    cwd,
	}).Info("loaded configuration")

	if *printDB {
		db := NewDatabase(config)
		if err := db.PrintDB(); err != nil {
			log.WithError(err).Fatal("unable to print database")
			os.Exit(1)
		}

		os.Exit(0)
	}

	if *refresh {
		db := NewDatabase(config)
		client := NewClient()
		lb := NewLogbook(config, db, client)

		auth, err := lb.Refresh()
		if err != nil {
			log.WithError(err).Fatal("cannot refresh")
		}

		if err := db.SetAuth(auth.Token, auth.Refresh); err != nil {
			log.WithField("auth", auth).WithError(err).Fatal("cannot save refresh token")
		}

		log.WithField("auth", auth).Info("saved new tokens")

		os.Exit(0)
	}

	if *auth != "" {
		splitted := strings.Split(*auth, ":")
		if len(splitted) != 2 {
			log.WithField("split", splitted).Fatal("unable to parse tokens")
		}

		db := NewDatabase(config)
		if err := db.SetAuth(splitted[0], splitted[1]); err != nil {
			log.WithError(err).Fatal("unable to save tokens")
		}

		log.Info("saved tokens")

		os.Exit(0)
	}

	return config
}
