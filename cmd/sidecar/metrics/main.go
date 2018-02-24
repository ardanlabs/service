package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/service/cmd/sidecar/metrics/collectors/expvar"
	"github.com/ardanlabs/service/cmd/sidecar/metrics/publishers/console"
	"github.com/ardanlabs/service/internal/platform/cfg"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func main() {

	// ============================================================
	// Configuration

	c, err := cfg.New(cfg.EnvProvider{Namespace: "METRICS"})
	if err != nil {
		log.Printf("%s. All config defaults in use.", err)
	}
	apiHost, err := c.String("API_HOST")
	if err != nil {
		apiHost = ":3000"
	}
	interval, err := c.Duration("INTERVAL")
	if err != nil {
		interval = 5 * time.Second
	}

	log.Printf("%s=%v", "API_HOST", apiHost)
	log.Printf("%s=%v", "INTERVAL", interval)

	// ============================================================
	// Start collectors and publishers

	expvar, err := expvar.New(apiHost)
	if err != nil {
		log.Fatalf("startup : Starting expvar collector : %v", err)
	}

	console, err := console.New(expvar, interval)
	if err != nil {
		log.Fatalf("startup : Starting console publisher : %v", err)
	}

	// ============================================================
	// Shutdown

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)
	<-osSignals

	log.Println("main : Start shutdown...")

	// ============================================================
	// Stop publishers

	console.Stop()

	log.Println("main : Completed")
}
