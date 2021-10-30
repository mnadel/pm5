package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type AdminServer struct {
	central *Central
	config  *Configuration
	db      *Database
	server  *http.Server
	mux     *mux.Router
}

func NewAdminServer(config *Configuration, db *Database, central *Central) *AdminServer {
	return &AdminServer{
		db:      db,
		config:  config,
		central: central,
		server: &http.Server{
			Addr:         config.AdminConsolePort,
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
		},
	}
}

func (as *AdminServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	as.server.Shutdown(ctx)
}

func (as *AdminServer) Start() {
	as.server.Handler = NewRoutes(as)

	log.WithField("port", as.config.AdminConsolePort).Info("starting console")
	log.Error(as.server.ListenAndServe())
}
