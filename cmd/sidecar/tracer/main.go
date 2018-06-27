package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/service/cmd/sidecar/tracer/handlers"
	"github.com/ardanlabs/service/internal/platform/cfg"
)

func main() {

	// =========================================================================
	// Logging

	log := log.New(os.Stdout, "TRACER : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	defer log.Println("main : Completed")

	// =========================================================================
	// Configuration

	c, err := cfg.New(cfg.EnvProvider{Namespace: "TRACER"})
	if err != nil {
		log.Printf("config : %s. All config defaults in use.", err)
	}
	readTimeout, err := c.Duration("READ_TIMEOUT")
	if err != nil {
		readTimeout = 5 * time.Second
	}
	writeTimeout, err := c.Duration("WRITE_TIMEOUT")
	if err != nil {
		writeTimeout = 5 * time.Second
	}
	shutdownTimeout, err := c.Duration("SHUTDOWN_TIMEOUT")
	if err != nil {
		shutdownTimeout = 5 * time.Second
	}
	apiHost, err := c.String("API_HOST")
	if err != nil {
		apiHost = "0.0.0.0:3002"
	}
	debugHost, err := c.String("DEBUG_HOST")
	if err != nil {
		debugHost = "0.0.0.0:4002"
	}
	zipkinHost, err := c.String("ZIPKIN_HOST")
	if err != nil {
		zipkinHost = "http://zipkin:9411/api/v2/spans"
	}

	log.Printf("config : %s=%v", "READ_TIMEOUT", readTimeout)
	log.Printf("config : %s=%v", "WRITE_TIMEOUT", writeTimeout)
	log.Printf("config : %s=%v", "SHUTDOWN_TIMEOUT", shutdownTimeout)
	log.Printf("config : %s=%v", "API_HOST", apiHost)
	log.Printf("config : %s=%v", "DEBUG_HOST", debugHost)
	log.Printf("config : %s=%v", "ZIPKIN_HOST", zipkinHost)

	// =========================================================================
	// Start Debug Service

	// /debug/pprof - Added to the default mux by the net/http/pprof package.

	debug := http.Server{
		Addr:           debugHost,
		Handler:        http.DefaultServeMux,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	// Not concerned with shutting this down when the
	// application is being shutdown.
	go func() {
		log.Printf("main : Debug Listening %s", debugHost)
		log.Printf("main : Debug Listener closed : %v", debug.ListenAndServe())
	}()

	// =========================================================================
	// Start API Service

	api := http.Server{
		Addr:           apiHost,
		Handler:        handlers.API(log, zipkinHost, apiHost),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		log.Printf("main : API Listening %s", apiHost)
		serverErrors <- api.ListenAndServe()
	}()

	// =========================================================================
	// Shutdown

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	// =========================================================================
	// Stop API Service

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		log.Fatalf("main : Error starting server: %v", err)

	case <-osSignals:
		log.Println("main : Start shutdown...")

		// Create context for Shutdown call.
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and load shed.
		if err := api.Shutdown(ctx); err != nil {
			log.Printf("main : Graceful shutdown did not complete in %v : %v", shutdownTimeout, err)
			if err := api.Close(); err != nil {
				log.Fatalf("main : Could not stop http server: %v", err)
			}
		}
	}
}
