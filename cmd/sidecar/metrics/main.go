package main

import (
	"encoding/json"
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
	"github.com/kelseyhightower/envconfig"
)

func main() {

	// =========================================================================
	// Logging

	log := log.New(os.Stdout, "TRACER : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	defer log.Println("main : Completed")

	// =========================================================================
	// Configuration

	var cfg struct {
		Web struct {
			DebugHost       string        `default:"0.0.0.0:4001" envconfig:"DEBUG_HOST"`
			ReadTimeout     time.Duration `default:"5s" envconfig:"READ_TIMEOUT"`
			WriteTimeout    time.Duration `default:"5s" envconfig:"WRITE_TIMEOUT"`
			ShutdownTimeout time.Duration `default:"5s" envconfig:"SHUTDOWN_TIMEOUT"`
		}
		Expvar struct {
			Host            string        `default:"0.0.0.0:3001" envconfig:"HOST"`
			Route           string        `default:"/metrics" envconfig:"ROUTE"`
			ReadTimeout     time.Duration `default:"5s" envconfig:"READ_TIMEOUT"`
			WriteTimeout    time.Duration `default:"5s" envconfig:"WRITE_TIMEOUT"`
			ShutdownTimeout time.Duration `default:"5s" envconfig:"SHUTDOWN_TIMEOUT"`
		}
		Collect struct {
			From string `default:"http://crud:4000/debug/vars" envconfig:"FROM"`
		}
		Publish struct {
			To       string        `default:"console" envconfig:"TO"`
			Interval time.Duration `default:"5s" envconfig:"INTERVAL"`
		}
	}

	if err := envconfig.Process("METRICS", &cfg); err != nil {
		log.Fatalf("main : Parsing Config : %v", err)
	}

	cfgJSON, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		log.Fatalf("main : Marshalling Config to JSON : %v", err)
	}
	log.Printf("config : %v\n", string(cfgJSON))

	// =========================================================================
	// Start Debug Service

	// /debug/pprof - Added to the default mux by the net/http/pprof package.

	debug := http.Server{
		Addr:           cfg.Web.DebugHost,
		Handler:        http.DefaultServeMux,
		ReadTimeout:    cfg.Web.ReadTimeout,
		WriteTimeout:   cfg.Web.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	// Not concerned with shutting this down when the
	// application is being shutdown.
	go func() {
		log.Printf("main : Debug Listening %s", cfg.Web.DebugHost)
		log.Printf("main : Debug Listener closed : %v", debug.ListenAndServe())
	}()

	// =========================================================================
	// Start expvar Service

	exp := expvar.New(log, cfg.Expvar.Host, cfg.Expvar.Route, cfg.Expvar.ReadTimeout, cfg.Expvar.WriteTimeout)
	defer exp.Stop(cfg.Expvar.ShutdownTimeout)

	// =========================================================================
	// Start collectors and publishers

	// Initalize to allow for the collection of metrics.
	collector, err := collector.New(cfg.Collect.From)
	if err != nil {
		log.Fatalf("main : Starting collector : %v", err)
	}

	// Create a stdout publisher.
	// TODO: Respect the cfg.publish.to config option.
	stdout := publisher.NewStdout(log)

	// Start the publisher to collect/publish metrics.
	publish, err := publisher.New(log, collector, cfg.Publish.Interval, exp.Publish, stdout.Publish)
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
