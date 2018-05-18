package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/service/cmd/sidecar/metrics/collector"
	"github.com/ardanlabs/service/cmd/sidecar/metrics/publisher"
	"github.com/ardanlabs/service/cmd/sidecar/metrics/publisher/expvar"
	"github.com/ardanlabs/service/internal/platform/cfg"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func main() {
	defer log.Println("main : Completed")

	c, err := cfg.New(cfg.EnvProvider{Namespace: "METRICS"})
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
	expHost, err := c.String("EXPVAR_HOST")
	if err != nil {
		expHost = "0.0.0.0:3001"
	}
	expRoute, err := c.String("EXPVAR_ROUTE")
	if err != nil {
		expRoute = "/metrics"
	}
	debugHost, err := c.String("DEBUG_HOST")
	if err != nil {
		debugHost = "0.0.0.0:4001"
	}
	crudHost, err := c.String("CRUD_HOST")
	if err != nil {
		crudHost = "http://crud:4000/debug/vars"
	}
	publishTo, err := c.String("PUBLISHER")
	if err != nil {
		publishTo = "console"
	}
	interval, err := c.Duration("INTERVAL")
	if err != nil {
		interval = 5 * time.Second
	}
	shutdownTimeout, err := c.Duration("SHUTDOWN_TIMEOUT")
	if err != nil {
		shutdownTimeout = 5 * time.Second
	}

	log.Printf("config : %s=%v", "READ_TIMEOUT", readTimeout)
	log.Printf("config : %s=%v", "WRITE_TIMEOUT", writeTimeout)
	log.Printf("config : %s=%v", "DEBUG_HOST", debugHost)
	log.Printf("config : %s=%v", "EXPVAR_HOST", expHost)
	log.Printf("config : %s=%v", "EXPVAR_ROUTE", expRoute)
	log.Printf("config : %s=%v", "CRUD_HOST", crudHost)
	log.Printf("config : %s=%v", "PUBLISHER", publishTo)
	log.Printf("config : %s=%v", "INTERVAL", interval)
	log.Printf("config : %s=%v", "SHUTDOWN_TIMEOUT", shutdownTimeout)

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
	// Start expvar Service

	exp := expvar.New(expHost, expRoute, readTimeout, writeTimeout)
	defer exp.Stop(shutdownTimeout)

	// =========================================================================
	// Start collectors and publishers

	// Initalize to allow for the collection of metrics.
	collector, err := collector.New(crudHost)
	if err != nil {
		log.Fatalf("main : Starting collector : %v", err)
	}

	// Start the publisher to collect/publish metrics.
	publish, err := publisher.New(collector, interval, exp.Publish, publisher.Stdout)
	if err != nil {
		log.Fatalf("main : Starting publisher : %v", err)
	}
	defer publish.Stop()

	// =========================================================================
	// Shutdown

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)
	<-osSignals

	log.Println("main : Start shutdown...")
}
