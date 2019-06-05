package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/service/cmd/sidecar/tracer/handlers"
	"github.com/ardanlabs/service/internal/platform/conf"
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
			APIHost         string        `conf:"default:0.0.0.0:3002"`
			DebugHost       string        `conf:"default:0.0.0.0:4002"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:5s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
		Zipkin struct {
			Host string `conf:"default:http://zipkin:9411/api/v2/spans"`
		}
	}

	if err := conf.Parse(os.Args[1:], "TRACER", &cfg); err != nil {
		if err == conf.ErrHelpWanted {
			usage, err := conf.Usage(&cfg)
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
	// Start API Service

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	api := http.Server{
		Addr:           cfg.Web.APIHost,
		Handler:        handlers.API(shutdown, log, cfg.Zipkin.Host, cfg.Web.APIHost),
		ReadTimeout:    cfg.Web.ReadTimeout,
		WriteTimeout:   cfg.Web.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		log.Printf("main : API Listening %s", cfg.Web.APIHost)
		serverErrors <- api.ListenAndServe()
	}()

	// =========================================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		log.Fatalf("main : Error starting server: %v", err)

	case sig := <-shutdown:
		log.Printf("main : %v : Start shutdown..", sig)

		// Create context for Shutdown call.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and load shed.
		err := api.Shutdown(ctx)
		if err != nil {
			log.Printf("main : Graceful shutdown did not complete in %v : %v", cfg.Web.ShutdownTimeout, err)
			err = api.Close()
		}

		// Log the status of this shutdown.
		switch {
		case sig == syscall.SIGSTOP:
			log.Fatal("main : Integrity issue caused shutdown")
		case err != nil:
			log.Fatalf("main : Could not stop server gracefully : %v", err)
		}
	}
}
