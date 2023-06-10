// Package expvar manages the publishing of metrics to stdout.
package expvar

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ardanlabs/service/foundation/logger"
	"github.com/dimfeld/httptreemux/v5"
)

// Expvar provide our basic publishing.
type Expvar struct {
	log    *logger.Logger
	server http.Server
	data   map[string]any
	mu     sync.Mutex
}

// New starts a service for consuming the raw expvar stats.
func New(log *logger.Logger, host string, route string, readTimeout, writeTimeout time.Duration, idleTimeout time.Duration) *Expvar {
	mux := httptreemux.New()
	exp := Expvar{
		log: log,
		server: http.Server{
			Addr:         host,
			Handler:      mux,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
			ErrorLog:     logger.NewStdLogger(log, logger.LevelError),
		},
	}

	mux.Handle("GET", route, exp.handler)

	go func() {
		log.Info("expvar", "status", "API listening", "host", host)
		if err := exp.server.ListenAndServe(); err != nil {
			log.Error("expvar", "ERROR", err)
		}
	}()

	return &exp
}

// Stop shuts down the service.
func (exp *Expvar) Stop(shutdownTimeout time.Duration) {
	exp.log.Info("expvar", "status", "start shutdown...")
	defer exp.log.Info("expvar: Completed")

	// Create context for Shutdown call.
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Asking listener to shut down and load shed.
	if err := exp.server.Shutdown(ctx); err != nil {
		exp.log.Error("expvar", "status", "graceful shutdown did not complete", "ERROR", err, "shutdownTimeout", shutdownTimeout)
		if err := exp.server.Close(); err != nil {
			exp.log.Error("expvar", "status", "could not stop http server", "ERROR", err)
		}
	}
}

// Publish is called by the publisher goroutine and saves the raw stats.
func (exp *Expvar) Publish(data map[string]any) {
	exp.mu.Lock()
	{
		exp.data = data
	}
	exp.mu.Unlock()
}

// handler is what consumers call to get the raw stats.
func (exp *Expvar) handler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	var data map[string]any
	exp.mu.Lock()
	{
		data = exp.data
	}
	exp.mu.Unlock()

	if err := json.NewEncoder(w).Encode(data); err != nil {
		exp.log.Error("expvar", "status", "encoding data", "ERROR", err)
	}

	log.Printf("expvar : (%d) : %s %s -> %s",
		http.StatusOK,
		r.Method, r.URL.Path,
		r.RemoteAddr,
	)
}
