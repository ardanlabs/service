package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/service/cmd/sidecar/metrics/collector"
	"github.com/ardanlabs/service/cmd/sidecar/metrics/publisher"
	"github.com/ardanlabs/service/internal/platform/cfg"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func main() {

	// =========================================================================
	// Configuration

	c, err := cfg.New(cfg.EnvProvider{Namespace: "METRICS"})
	if err != nil {
		log.Printf("%s. All config defaults in use.", err)
	}
	apiHost, err := c.String("API_HOST")
	if err != nil {
		apiHost = "http://localhost:4000/debug/vars"
	}
	interval, err := c.Duration("INTERVAL")
	if err != nil {
		interval = 5 * time.Second
	}
	publishTo, err := c.String("PUBLISHER")
	if err != nil {
		publishTo = "console"
	}

	log.Printf("%s=%v", "API_HOST", apiHost)
	log.Printf("%s=%v", "INTERVAL", interval)
	log.Printf("%s=%v", "PUBLISHER", publishTo)

	// =========================================================================
	// Start collectors and publishers

	// Initalize to allow for the collection of metrics.
	expvar, err := expvar.New(apiHost)
	if err != nil {
		log.Fatalf("startup : Starting collector : %v", err)
	}

	// Determine which publisher to use.
	f := publisher.Console
	switch publishTo {
	case publisher.TypeDatadog:
		f = publisher.Datadog
	}

	// Start the publisher to collect/publish metrics.
	publish, err := publisher.New(expvar, f, interval)
	if err != nil {
		log.Fatalf("startup : Starting publisher : %v", err)
	}

	// =========================================================================
	// Shutdown

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)
	<-osSignals

	log.Println("main : Start shutdown...")
	defer log.Println("main : Completed")

	// =========================================================================
	// Stop publishers

	publish.Stop()
}
