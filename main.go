package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
)

const (
	ERR_BLEDEADLOCK  = 37
	ERR_NODISCONNECT = 41
	ERR_USAGE        = 43
	ERR_CANTSTORE    = 47
	ERR_NOWORKOUT    = 53
)

func main() {
	var auth bool
	var access string
	var refresh string

	flag.BoolVar(&auth, "auth", false, "generate a url to get an auth token")
	flag.StringVar(&access, "access", "", "set the access token")
	flag.StringVar(&refresh, "refresh", "", "set the refresh token")
	flag.Parse()

	if !valid(auth, access, refresh) {
		flag.PrintDefaults()
		os.Exit(ERR_USAGE)
	} else if auth {
		lb, err := NewLogbook(NewConfiguration())
		if err != nil {
			log.WithError(err).Fatal("cannot generate auth url")
		}
		defer lb.Close()
		lb.Authenticate()
		os.Exit(0)
	} else if access != "" && refresh != "" {
		lb, err := NewLogbook(NewConfiguration())
		if err != nil {
			log.WithError(err).Fatal("cannot store tokens")
		}
		defer lb.Close()
		if err := lb.SetTokens(access, refresh); err != nil {
			log.WithError(err).Fatal("error storing tokens")
			os.Exit(ERR_CANTSTORE)
		}
		os.Exit(0)
	}

	config := NewConfiguration()

	startAdminConsole(config)

	central := NewCentral(config)

	log.Info("starting central")
	central.Listen()

	log.Info("central exiting")
}

func valid(auth bool, access, refresh string) bool {
	if !auth && access == "" && refresh == "" {
		return true
	} else if auth {
		return access == "" && refresh == ""
	} else {
		return access != "" && refresh != ""
	}
}
