package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/conf"
	"github.com/ardanlabs/service/cmd/sidecar/metrics/internal/collector"
	"github.com/ardanlabs/service/cmd/sidecar/metrics/internal/publisher"
	"github.com/ardanlabs/service/cmd/sidecar/metrics/internal/publisher/expvar"
)

func main() {

	// =========================================================================
	// Logging

	log := log.New(os.Stdout, "METRICS : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	defer log.Println("main : Completed")

	// =========================================================================
	// Configuration

	var cfg struct {
		Web struct {
			DebugHost       string        `conf:"default:0.0.0.0:4001"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:5s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
		Expvar struct {
			Host            string        `conf:"default:0.0.0.0:3001"`
			Route           string        `conf:"default:/metrics"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:5s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
		Collect struct {
			From string `conf:"default:http://sales-api:4000/debug/vars"`
		}
		Publish struct {
			To       string        `conf:"default:console"`
			Interval time.Duration `conf:"default:5s"`
		}
	}

	if err := conf.Parse(os.Args[1:], "METRICS", &cfg); err != nil {
		if err == conf.ErrHelpWanted {
			usage, err := conf.Usage("METRICS", &cfg)
			if err != nil {
				log.Fatalf("main : Parsing Config : %v", err)
			}
			fmt.Println(usage)
			return
		}
		log.Fatalf("main : Parsing Config : %v", err)
	}

	out, err := conf.String(&cfg)
	if err != nil {
		log.Fatalf("main : Marshalling Config for output : %v", err)
	}
	log.Printf("main : Config :\n%v\n", out)

	// =========================================================================
	// Start Debug Service. Not concerned with shutting this down when the
	// application is being shutdown.
	//
	// /debug/pprof - Added to the default mux by the net/http/pprof package.
	go func() {
		log.Printf("main : Debug Listening %s", cfg.Web.DebugHost)
		log.Printf("main : Debug Listener closed : %v", http.ListenAndServe(cfg.Web.DebugHost, http.DefaultServeMux))
	}()

	// =========================================================================
	// Start expvar Service

	exp := expvar.New(log, cfg.Expvar.Host, cfg.Expvar.Route, cfg.Expvar.ReadTimeout, cfg.Expvar.WriteTimeout)
	defer exp.Stop(cfg.Expvar.ShutdownTimeout)

	// =========================================================================
	// Start collectors and publishers

	// Initialize to allow for the collection of metrics.
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
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown

	log.Println("main : Start shutdown...")
}
