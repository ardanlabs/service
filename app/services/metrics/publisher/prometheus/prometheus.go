// Package prometheus provides suppoert for sending metrics to prometheus.
package prometheus

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/dimfeld/httptreemux/v5"
	"golang.org/x/exp/slog"
)

// Exporter implements the prometheus exporter support.
type Exporter struct {
	log    *slog.Logger
	server http.Server
	data   map[string]any
	mu     sync.Mutex
}

// New constructs an Exporter for use.
func New(log *slog.Logger, host string, route string, readTimeout, writeTimeout time.Duration, idleTimeout time.Duration) *Exporter {
	mux := httptreemux.NewContextMux()

	exp := Exporter{
		log: log,
		server: http.Server{
			Addr:         host,
			Handler:      mux,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
	}

	mux.Handle(http.MethodGet, route, exp.handler)

	go func() {
		log.Info("prometheus", "status", "API listening", "host", host)

		if err := exp.server.ListenAndServe(); err != nil {
			log.Info("prometheus", "ERROR", err)
		}
	}()

	return &exp
}

// Publish stores a deep copy of the data for publishing.
func (exp *Exporter) Publish(data map[string]any) {
	exp.mu.Lock()
	defer exp.mu.Unlock()

	exp.data = deepCopyMap(data)
}

// Stop turns off all the prometheus support.
func (exp *Exporter) Stop(shutdownTimeout time.Duration) {
	exp.log.Info("prometheus", "status", "start shutdown...")
	defer exp.log.Info("prometheus: Completed")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := exp.server.Shutdown(ctx); err != nil {
		exp.log.Info("prometheus", "status", "graceful shutdown did not complete", "ERROR", err, "shutdownTimeout", shutdownTimeout)

		if err := exp.server.Close(); err != nil {
			exp.log.Info("prometheus", "status", "could not stop http server", "ERROR", err)
		}
	}
}

// =============================================================================

func (exp *Exporter) handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.WriteHeader(http.StatusOK)

	var data map[string]any
	exp.mu.Lock()
	{
		data = deepCopyMap(exp.data)
	}
	exp.mu.Unlock()

	out(w, "", data)

	exp.log.Info(fmt.Sprintf("prometheus: (%d) : %s %s -> %s",
		http.StatusOK,
		r.Method, r.URL.Path,
		r.RemoteAddr,
	))
}

// =============================================================================

func deepCopyMap(source map[string]any) map[string]any {
	result := make(map[string]any)

	for k, v := range source {
		switch vm := v.(type) {
		case map[string]any:
			result[k] = deepCopyMap(vm)

		case int64:
			result[k] = float64(vm)

		case float64:
			result[k] = vm

		case bool:
			result[k] = 0.0
			if vm {
				result[k] = 1.0
			}
		}
	}

	return result
}

func out(w io.Writer, prefix string, data map[string]any) {
	if prefix != "" {
		prefix += "_"
	}

	for k, v := range data {
		writeKey := fmt.Sprintf("%s%s", prefix, k)

		switch vm := v.(type) {
		case float64:
			fmt.Fprintf(w, "%s %.f\n", writeKey, vm)

		case map[string]any:
			out(w, writeKey, vm)

		default:
			// Discard this value.
		}
	}
}
