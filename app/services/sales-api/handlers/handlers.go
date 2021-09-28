// Package handlers manages the different versions of the API.
package handlers

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/service/app/services/sales-api/handlers/debug/checkgrp"
	v1 "github.com/ardanlabs/service/app/services/sales-api/handlers/v1"
	"github.com/ardanlabs/service/business/sys/auth"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// V1 asks for v1 related settings.
type V1 struct {
	APIHost   string
	DebugHost string
}

// Config helps to capture the necessary settings.
type Config struct {
	Build           string
	Log             *zap.SugaredLogger
	DB              *sqlx.DB
	Auth            *auth.Auth
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	V1              V1
}

// StartServers brings up the different versions of the services.
func StartServers(cfg Config) func() error {

	// Start the debug server.
	startDebugServer(cfg)

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the API servers.
	var servers []http.Server
	servers = append(servers, startV1Server(cfg, shutdown, serverErrors))

	// This function blocks the caller and handles the shutdown sequence.
	f := func() error {

		// This is a blocking select.
		select {
		case err := <-serverErrors:
			return fmt.Errorf("server error: %w", err)

		case sig := <-shutdown:
			cfg.Log.Infow("shutdown", "status", "shutdown started", "signal", sig)
			defer cfg.Log.Infow("shutdown", "status", "shutdown complete", "signal", sig)

			// Give outstanding requests a deadline for completion.
			ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
			defer cancel()

			g, ctx := errgroup.WithContext(ctx)

			// Shutdown each server in parallel.
			for _, server := range servers {
				server := server
				g.Go(func() error {
					if err := server.Shutdown(ctx); err != nil {
						server.Close()
						return fmt.Errorf("could not stop server gracefully: %w", err)
					}
					return nil
				})
			}

			// Wait for the G's to return an error. If one exists,
			// pass it to the caller.
			if err := g.Wait(); err != nil {
				return err
			}

			return nil
		}
	}

	return f
}

func startDebugServer(cfg Config) {
	cfg.Log.Infow("startup", "status", "debug v1 router started", "host", cfg.V1.DebugHost)

	// The Debug function returns a mux to listen and serve on for all the debug
	// related endpoints. This include the standard library endpoints.

	// Construct the mux for the debug calls.
	debugMux := debugMux(cfg.Build, cfg.Log, cfg.DB)

	// Start the service listening for debug requests.
	// Not concerned with shutting this down with load shedding.
	go func() {
		if err := http.ListenAndServe(cfg.V1.DebugHost, debugMux); err != nil {
			cfg.Log.Errorw("shutdown", "status", "debug v1 router closed", "host", cfg.V1.DebugHost, "ERROR", err)
		}
	}()
}

func startV1Server(cfg Config, shutdown chan os.Signal, serverErrors chan error) http.Server {
	cfg.Log.Infow("startup", "status", "initializing API support")

	// Construct the mux for the API calls.
	apiMux := v1.APIMux(v1.APIMuxConfig{
		Shutdown: shutdown,
		Log:      cfg.Log,
		Auth:     cfg.Auth,
		DB:       cfg.DB,
	})

	// Construct a server to service the requests against the mux.
	api := http.Server{
		Addr:         cfg.V1.APIHost,
		Handler:      apiMux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		ErrorLog:     zap.NewStdLog(cfg.Log.Desugar()),
	}

	// Start the service listening for api requests.
	go func() {
		cfg.Log.Infow("startup", "status", "v1 api router started", "host", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	return api
}

// debugStandardLibraryMux registers all the debug routes from the standard library
// into a new mux bypassing the use of the DefaultServerMux. Using the
// DefaultServerMux would be a security risk since a dependency could inject a
// handler into our service without us knowing it.
func debugStandardLibraryMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Register all the standard library debug endpoints.
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars", expvar.Handler())

	return mux
}

// debugMux registers all the debug standard library routes and then custom
// debug application routes for the service. This bypassing the use of the
// DefaultServerMux. Using the DefaultServerMux would be a security risk since
// a dependency could inject a handler into our service without us knowing it.
func debugMux(build string, log *zap.SugaredLogger, db *sqlx.DB) http.Handler {
	mux := debugStandardLibraryMux()

	// Register debug check endpoints.
	cgh := checkgrp.Handlers{
		Build: build,
		Log:   log,
		DB:    db,
	}
	mux.HandleFunc("/debug/readiness", cgh.Readiness)
	mux.HandleFunc("/debug/liveness", cgh.Liveness)

	return mux
}
