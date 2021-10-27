package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

const (
	MAX_LOGFILE_SIZE   = 10485760 // 10MB
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
		panic(err)
	}

	return &Configuration{
		DBFile: f.Name(),
	}
}

func NewConfiguration() *Configuration {
	printDB := flag.Bool("dump", false, "print the contents of the database")
	refresh := flag.Bool("refresh", false, "get a new refresh token")
	authURL := flag.Bool("authurl", false, "print the auth url")
	migrate := flag.Bool("migrate", false, "migrate db records")

	host := flag.String("host", "log.concept2.com", "specify the logbook service hostname")
	auth := flag.String("auth", "", "set the auth token in the form of id:secret")
	dbFile := flag.String("db", "pm5.boltdb", "path to db file")
	logFile := flag.String("logfile", "-", "path to logfile, - for stdout")
	logLevel := flag.String("loglevel", "info", "the logrus log level")
	port := flag.String("port", ":2112", "web console port")

	bleWatchdogWorkoutDeadline := flag.Duration("deadline", time.Minute*35, "max duration after we connect to the PM5 before we expect to receive a workout summary")
	bleWatchdogWorkoutDisconnect := flag.Duration("disconn", time.Minute*7, "max duration after workout sumary is received before we expect a disconnect")
	bleWatchdogDeadline := flag.Duration("scan", time.Second*60, "max duration between scans we'll tolerate")

	flag.Parse()

	if *authURL {
		printAuthURL(*host)
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

	configureLogger(*logLevel, *logFile)

	cwd, _ := os.Getwd()

	log.WithFields(log.Fields{
		"config": *config,
		"user":   os.Getenv("USER"),
		"cwd":    cwd,
	}).Info("loaded configuration")

	if *printDB {
		dumpDB(config)
		os.Exit(0)
	} else if *refresh {
		refreshTokens(config)
		os.Exit(0)
	} else if *auth != "" {
		saveAuth(*auth, config)
		os.Exit(0)
	} else if *migrate {
		migrateDB(config)
		os.Exit(0)
	}

	return config
}

func configureLogger(logLevel, logFile string) {
	parsedLogLevel, err := log.ParseLevel(logLevel)
	if err != nil {
		Panic(err, "cannot parse log level %s", logLevel)
	}

	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
		ForceQuote:    true,
	})

	if logFile == "-" {
		log.SetOutput(os.Stdout)
	} else {
		log.SetOutput(&lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    5, // megabytes
			MaxBackups: 10,
			MaxAge:     int(time.Hour.Hours() * 24 * 31),
			Compress:   false,
		})
	}

	log.SetLevel(parsedLogLevel)
	log.SetReportCaller(parsedLogLevel == log.DebugLevel)
}

func printAuthURL(host string) {
	fmt.Println("** please visit the below url **")
	uriFmt := "https://%s/oauth/authorize?client_id=%s&scope=results:write&response_type=code&redirect_uri=%s"
	fmt.Printf(uriFmt, host, PM5_OAUTH_APPID, PM5_OAUTH_CALLBACK)
	fmt.Println("")
}

func migrateDB(config *Configuration) {
	db := NewDatabase(config)
	migrator := NewDBMigrator(db)

	if err := migrator.Migrate(); err != nil {
		panic(err)
	}
}

func dumpDB(config *Configuration) {
	db := NewDatabase(config)
	if err := db.PrintDB(); err != nil {
		panic(err)
	}
}

func refreshTokens(config *Configuration) {
	db := NewDatabase(config)
	auth, err := db.GetAuth()
	if err != nil {
		Panic(err, "cannot find auth")
	}

	auth, err = RefreshAuth(config, NewClient(), auth)
	if err != nil {
		Panic(err, "cannot refresh auth")
	}

	if err := db.SetAuth(auth.Token, auth.Refresh); err != nil {
		Panic(err, "cannot save auth %v", auth)
	}

	log.WithField("auth", auth).Info("saved new tokens")
}

func saveAuth(auth string, config *Configuration) {
	splitted := strings.Split(auth, ":")
	if len(splitted) != 2 {
		Panic(fmt.Errorf("len=%d", len(splitted)), "cannot parse: %v")
	}

	db := NewDatabase(config)
	if err := db.SetAuth(splitted[0], splitted[1]); err != nil {
		Panic(err, "cannot save tokens")
	}

	log.Info("saved tokens")
}
