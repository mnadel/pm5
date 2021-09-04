package main

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func StartAdminConsole() {
	go func() {
		r := mux.NewRouter()

		r.Handle("/metrics", promhttp.Handler())

		r.HandleFunc("/ble/last-scan", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")

			c := make(chan prometheus.Metric, 1)
			MetricLastScan.Collect(c)
			m := dto.Metric{}
			_ = (<-c).Write(&m)

			t := time.Unix(int64(*m.Gauge.Value), 0)

			fmt.Fprint(w, t.Format(RFC8601))

			fmt.Fprint(w, "\n👍\n")
		})

		r.HandleFunc("/debug/block/{rate}", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")

			vars := mux.Vars(r)
			key := vars["rate"]
			rate := shouldParseAtoi(key)

			runtime.SetBlockProfileRate(rate)

			fmt.Fprint(w, "\n👍\n")
		})

		r.HandleFunc("/debug/mutex/{rate}", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")

			vars := mux.Vars(r)
			key := vars["rate"]
			rate := shouldParseAtoi(key)

			runtime.SetMutexProfileFraction(rate)

			currRate := runtime.SetMutexProfileFraction(-1)
			fmt.Fprint(w, currRate)
			fmt.Fprint(w, "\n👍\n")
		})

		r.HandleFunc("/debug/pprof/", pprof.Index)
		r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		r.HandleFunc("/debug/pprof/profile", pprof.Profile)
		r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

		// Manually add support for paths linked to by index page at /debug/pprof/
		r.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
		r.Handle("/debug/pprof/heap", pprof.Handler("heap"))
		r.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
		r.Handle("/debug/pprof/block", pprof.Handler("block"))

		log.WithField("port", Config.AdminConsolePort).Info("starting console")
		log.Info(http.ListenAndServe(Config.AdminConsolePort, r))
	}()
}
