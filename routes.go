package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gorilla/mux"
)

func NewRoutes(as *AdminServer) *mux.Router {
	m := mux.NewRouter()

	// public api

	uc := NewUserController(as.config, as.db)

	m.HandleFunc("/user", uc.HandleGetUser).Methods(http.MethodGet)
	m.HandleFunc("/user/{uuid}", uc.HandleSetUser).Methods(http.MethodPut)

	// internal api

	m.Handle("/metrics", promhttp.Handler())

	m.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		if j, err := json.Marshal(as.central.Stats()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			fmt.Fprint(w, string(j))
		}
	})

	m.HandleFunc("/ble/lastscan", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		val := GetPromGaugeValue(&MetricLastBLEScan)
		t := time.Unix(int64(val), 0)

		fmt.Fprint(w, t.Format(ISO8601))
		fmt.Fprint(w, "\n👍\n")
	})

	m.HandleFunc("/debug/block/{rate}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		vars := mux.Vars(r)
		key := vars["rate"]
		rate := ShouldParseAtoi(key)

		runtime.SetBlockProfileRate(rate)

		fmt.Fprint(w, "\n👍\n")
	})

	m.HandleFunc("/debug/mutex/{rate}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		vars := mux.Vars(r)
		key := vars["rate"]
		rate := ShouldParseAtoi(key)

		runtime.SetMutexProfileFraction(rate)

		currRate := runtime.SetMutexProfileFraction(-1)
		fmt.Fprint(w, currRate)
		fmt.Fprint(w, "\n👍\n")
	})

	m.HandleFunc("/debug/pprof/", pprof.Index)
	m.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	m.HandleFunc("/debug/pprof/profile", pprof.Profile)
	m.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	// Manually add support for paths linked to by index page at /debug/pprof/
	m.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	m.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	m.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	m.Handle("/debug/pprof/block", pprof.Handler("block"))

	return m
}
