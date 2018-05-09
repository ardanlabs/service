package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/service/cmd/sidecar/metrics/collector"
	"github.com/ardanlabs/service/cmd/sidecar/metrics/publisher"
	"github.com/ardanlabs/service/internal/platform/cfg"
)

/*
	Need to add the debug route with default mux.
	Add the health checks.
	Let's have expvarparmon hit this service instead.
*/

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func main() {

	// =========================================================================
	// Configuration

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
	apiHost, err := c.String("API_HOST")
	if err != nil {
		apiHost = "http://crud:4000/debug/vars"
	}
	interval, err := c.Duration("INTERVAL")
	if err != nil {
		interval = 5 * time.Second
	}
	publishTo, err := c.String("PUBLISHER")
	if err != nil {
		publishTo = "console"
	}
	dataDogAPIKey, err := c.String("DATADOG_APIKEY")
	if err != nil {
		dataDogAPIKey = "03f53bb094715f2eb8ac843c90c00232"
	}
	dataDogHost, err := c.String("DATADOG_HOST")
	if err != nil {
		dataDogHost = "https://app.datadoghq.com/api/v1/series"
	}
	debugHost, err := c.String("DEBUG_HOST")
	if err != nil {
		debugHost = "0.0.0.0:4001"
	}

	log.Printf("config : %s=%v", "READ_TIMEOUT", readTimeout)
	log.Printf("config : %s=%v", "WRITE_TIMEOUT", writeTimeout)
	log.Printf("config : %s=%v", "API_HOST", apiHost)
	log.Printf("config : %s=%v", "INTERVAL", interval)
	log.Printf("config : %s=%v", "PUBLISHER", publishTo)
	log.Printf("config : %s=%v", "DATADOG_APIKEY", dataDogAPIKey)
	log.Printf("config : %s=%v", "DATADOG_HOST", dataDogHost)
	log.Printf("config : %s=%v", "DEBUG_HOST", debugHost)

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
	// Start collectors and publishers

	// Initalize to allow for the collection of metrics.
	expvar, err := expvar.New(apiHost)
	if err != nil {
		log.Fatalf("main : Starting collector : %v", err)
	}

	// Determine which publisher to use.
	f := publisher.Console
	switch publishTo {
	case publisher.TypeConsole:
		log.Println("config : PUB_TYPE=Console")

	case publisher.TypeDatadog:
		log.Println("config : PUB_TYPE=Datadog")
		d := publisher.NewDatadog(dataDogAPIKey, dataDogHost)
		f = d.Publish

	default:
		log.Fatalln("main : No publisher provided, using Console.")
	}

	// Start the publisher to collect/publish metrics.
	publish, err := publisher.New(expvar, f, interval)
	if err != nil {
		log.Fatalf("main : Starting publisher : %v", err)
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
