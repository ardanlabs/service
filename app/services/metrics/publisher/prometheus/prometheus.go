package prometheus

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/dimfeld/httptreemux/v5"
	"go.uber.org/zap"
)

type Exporter struct {
	log    *zap.SugaredLogger
	server http.Server
	data   map[string]any
	mu     sync.Mutex
}

func New(log *zap.SugaredLogger, host string, route string, readTimeout, writeTimeout time.Duration, idleTimeout time.Duration) *Exporter {
	mux := httptreemux.New()
	exp := Exporter{
		log: log,
		server: http.Server{
			Addr:         host,
			Handler:      mux,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
			ErrorLog:     zap.NewStdLog(log.Desugar()),
		},
	}

	mux.Handle("GET", route, exp.handler)

	go func() {
		log.Infow("prometheus", "status", "API listening", "host", host)
		if err := exp.server.ListenAndServe(); err != nil {
			log.Errorw("ERROR", zap.Error(err))
		}
	}()

	return &exp
}

func (exp *Exporter) handler(w http.ResponseWriter, r *http.Request, m map[string]string) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.WriteHeader(http.StatusOK)

	var data map[string]any
	exp.mu.Lock()
	{
		data = deepCopyMap(exp.data)
	}
	exp.mu.Unlock()

	out(w, "", data)

	log.Printf("prometheus: (%d) : %s %s -> %s",
		http.StatusOK,
		r.Method, r.URL.Path,
		r.RemoteAddr,
	)
}

func (exp *Exporter) Publish(data map[string]any) {
	exp.mu.Lock()
	{
		exp.data = deepCopyMap(data)
	}
	exp.mu.Unlock()
}

func (exp *Exporter) Stop(shutdownTimeout time.Duration) {
	exp.log.Infow("prometheus", "status", "start shutdown...")
	defer exp.log.Infow("prometheus: Completed")

	// Create context for Shutdown call.
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Asking listener to shut down and load shed.
	if err := exp.server.Shutdown(ctx); err != nil {
		exp.log.Errorw("prometheus", "status", "graceful shutdown did not complete", "ERROR", err, "shutdownTimeout", shutdownTimeout)
		if err := exp.server.Close(); err != nil {
			exp.log.Errorw("prometheus", "status", "could not stop http server", "ERROR", err)
		}
	}
}

func deepCopyMap(source map[string]any) map[string]any {
	result := make(map[string]any)
	for k, v := range source {
		vm, ok := v.(map[string]any)
		if ok {
			result[k] = deepCopyMap(vm)
		} else {
			val, err := toFloat64(v)
			if err != nil {
				continue
			}

			result[k] = val
		}
	}

	return result
}

func toFloat64(v interface{}) (float64, error) {
	switch v := v.(type) {
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	}
	return 0, fmt.Errorf("unexpected value type: %#v", v)
}

func out(w io.Writer, prefix string, data map[string]any) {
	if prefix != "" {
		prefix += "_"
	}

	for key, val := range data {
		writeKey := fmt.Sprintf("%s%s", prefix, key)
		switch val.(type) {
		case float64:
			fmt.Fprintf(w, fmt.Sprintf("%s %.f\n", writeKey, val))
		case map[string]any:
			out(w, writeKey, val.(map[string]any))
		default:
			// Discard stuff
		}
	}
}
