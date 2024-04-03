package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ardanlabs/conf/v3"
	"github.com/ardanlabs/service/apis/services/salesapiweb/build/all"
	"github.com/ardanlabs/service/apis/services/salesapiweb/build/crud"
	"github.com/ardanlabs/service/apis/services/salesapiweb/build/reporting"
	"github.com/ardanlabs/service/app/service"
	"github.com/ardanlabs/service/business/api/debug"
	"github.com/ardanlabs/service/business/api/mux"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
)

/*
	Need to figure out timeouts for http service.
*/

var build = "develop"
var routes = "all" // go build -ldflags "-X main.routes=crud"

func main() {
	var log *logger.Logger

	events := logger.Events{
		Error: func(ctx context.Context, r logger.Record) {
			log.Info(ctx, "******* SEND ALERT ******")
		},
	}

	traceIDFn := func(ctx context.Context) string {
		return web.GetTraceID(ctx)
	}

	log = logger.NewWithEvents(os.Stdout, logger.LevelInfo, "SALES-API", traceIDFn, events)

	// -------------------------------------------------------------------------

	ctx := context.Background()

	if err := run(ctx, log); err != nil {
		log.Error(ctx, "startup", "msg", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *logger.Logger) error {

	// -------------------------------------------------------------------------
	// Configuration

	cfg := service.Config{
		Version: conf.Version{
			Build: build,
			Desc:  "Service Project",
		},
	}

	const prefix = "SALES"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// -------------------------------------------------------------------------
	// Construct Service

	service, err := service.New(ctx, log, cfg)
	if err != nil {
		return fmt.Errorf("constructing service: %w", err)
	}

	// -------------------------------------------------------------------------
	// Start Debug Service

	go func() {
		log.Info(ctx, "startup", "status", "debug v1 router started", "host", cfg.Web.DebugHost)

		if err := http.ListenAndServe(cfg.Web.DebugHost, debug.Mux()); err != nil {
			log.Error(ctx, "shutdown", "status", "debug v1 router closed", "host", cfg.Web.DebugHost, "msg", err)
		}
	}()

	// -------------------------------------------------------------------------
	// Start API Service

	log.Info(ctx, "startup", "status", "initializing V1 API support")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	cfgMux := mux.Config{
		Build:    build,
		Shutdown: shutdown,
		Log:      log,
		Auth:     service.Auth,
		DB:       service.DB,
		Tracer:   service.Tracer,
		BusCrud: mux.BusCrud{
			Delegate: service.Bus.Crud.Delegate,
			Home:     service.Bus.Crud.Home,
			Product:  service.Bus.Crud.Product,
			User:     service.Bus.Crud.User,
		},
		BusView: mux.BusView{
			Product: service.Bus.View.Product,
		},
	}

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      mux.WebAPI(cfgMux, buildRoutes(), mux.WithCORS(cfg.Web.CORSAllowedOrigins)),
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     logger.NewStdLogger(log, logger.LevelError),
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Info(ctx, "startup", "status", "api router started", "host", api.Addr)

		serverErrors <- api.ListenAndServe()
	}()

	// -------------------------------------------------------------------------
	// Shutdown

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Info(ctx, "shutdown", "status", "shutdown started", "signal", sig)
		defer log.Info(ctx, "shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(ctx, cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}

func buildRoutes() mux.RouteAdder {

	// The idea here is that we can build different versions of the binary
	// with different sets of exposed web APIs. By default we build a single
	// an instance with all the web APIs.
	//
	// Here is the scenario. It would be nice to build two binaries, one for the
	// transactional APIs (CRUD) and one for the reporting APIs. This would allow
	// the system to run two instances of the database. One instance tuned for the
	// transactional database calls and the other tuned for the reporting calls.
	// Tuning meaning indexing and memory requirements. The two databases can be
	// kept in sync with replication.

	switch routes {
	case "crud":
		return crud.Routes()

	case "reporting":
		return reporting.Routes()
	}

	return all.Routes()
}
