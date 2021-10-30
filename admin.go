package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

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
	mux := mux.NewRouter()

	return &AdminServer{
		db:      db,
		mux:     mux,
		config:  config,
		central: central,
		server: &http.Server{
			Addr:         config.AdminConsolePort,
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
			Handler:      mux,
		},
	}
}

func (as *AdminServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	as.server.Shutdown(ctx)
}

func (as *AdminServer) Start() {
	// public api

	uc := NewUserController(as.config, as.db)

	as.mux.HandleFunc("/user", uc.HandleLoginQuery).Methods(http.MethodGet)
	as.mux.HandleFunc("/user/{uuid}", uc.HandleLogin).Methods(http.MethodPut)

	// internal api

	as.mux.Handle("/metrics", promhttp.Handler())

	as.mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		if j, err := json.Marshal(as.central.Stats()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			fmt.Fprint(w, string(j))
		}
	})

	as.mux.HandleFunc("/ble/lastscan", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		val := GetPromGaugeValue(&MetricLastBLEScan)
		t := time.Unix(int64(val), 0)

		fmt.Fprint(w, t.Format(ISO8601))
		fmt.Fprint(w, "\n👍\n")
	})

	as.mux.HandleFunc("/debug/block/{rate}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		vars := mux.Vars(r)
		key := vars["rate"]
		rate := ShouldParseAtoi(key)

		runtime.SetBlockProfileRate(rate)

		fmt.Fprint(w, "\n👍\n")
	})

	as.mux.HandleFunc("/debug/mutex/{rate}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		vars := mux.Vars(r)
		key := vars["rate"]
		rate := ShouldParseAtoi(key)

		runtime.SetMutexProfileFraction(rate)

		currRate := runtime.SetMutexProfileFraction(-1)
		fmt.Fprint(w, currRate)
		fmt.Fprint(w, "\n👍\n")
	})

	as.mux.HandleFunc("/debug/pprof/", pprof.Index)
	as.mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	as.mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	as.mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	// Manually add support for paths linked to by index page at /debug/pprof/
	as.mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	as.mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	as.mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	as.mux.Handle("/debug/pprof/block", pprof.Handler("block"))

	log.WithField("port", as.config.AdminConsolePort).Info("starting console")
	log.Error(as.server.ListenAndServe())
}
