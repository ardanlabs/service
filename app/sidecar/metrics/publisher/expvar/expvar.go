package expvar

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/dimfeld/httptreemux/v5"
)

// Expvar provide our basic publishing.
type Expvar struct {
	log    *log.Logger
	server http.Server
	data   map[string]interface{}
	mu     sync.Mutex
}

// New starts a service for consuming the raw expvar stats.
func New(log *log.Logger, host string, route string, readTimeout, writeTimeout time.Duration) *Expvar {
	mux := httptreemux.New()
	exp := Expvar{
		log: log,
		server: http.Server{
			Addr:           host,
			Handler:        mux,
			ReadTimeout:    readTimeout,
			WriteTimeout:   writeTimeout,
			MaxHeaderBytes: 1 << 20,
		},
	}

	mux.Handle("GET", route, exp.handler)

	go func() {
		log.Println("expvar : API Listening", host)
		if err := exp.server.ListenAndServe(); err != nil {
			log.Println("expvar : ERROR :", err)
		}
	}()

	return &exp
}

// Stop shuts down the service.
func (exp *Expvar) Stop(shutdownTimeout time.Duration) {
	exp.log.Println("expvar : Start shutdown...")
	defer exp.log.Println("expvar : Completed")

	// Create context for Shutdown call.
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Asking listener to shutdown and load shed.
	if err := exp.server.Shutdown(ctx); err != nil {
		exp.log.Printf("expvar : Graceful shutdown did not complete in %v : %v", shutdownTimeout, err)
		if err := exp.server.Close(); err != nil {
			exp.log.Fatalf("expvar : Could not stop http server: %v", err)
		}
	}
}

// Publish is called by the publisher goroutine and saves the raw stats.
func (exp *Expvar) Publish(data map[string]interface{}) {
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

	var data map[string]interface{}
	exp.mu.Lock()
	{
		data = exp.data
	}
	exp.mu.Unlock()

	if err := json.NewEncoder(w).Encode(data); err != nil {
		exp.log.Println("expvar : ERROR :", err)
	}

	log.Printf("expvar : (%d) : %s %s -> %s",
		http.StatusOK,
		r.Method, r.URL.Path,
		r.RemoteAddr,
	)
}
