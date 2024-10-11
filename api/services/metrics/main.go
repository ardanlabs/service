package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/ardanlabs/service/api/services/metrics/collector"
	"github.com/ardanlabs/service/api/services/metrics/publisher"
	expvarsrv "github.com/ardanlabs/service/api/services/metrics/publisher/expvar"
	prometheussrv "github.com/ardanlabs/service/api/services/metrics/publisher/prometheus"
	"github.com/ardanlabs/service/app/sdk/debug"
	"github.com/ardanlabs/service/foundation/logger"
)

var build = "develop"

func main() {
	var log *logger.Logger

	events := logger.Events{
		Error: func(ctx context.Context, r logger.Record) { log.Info(ctx, "******* SEND ALERT ******") },
	}

	traceIDFn := func(ctx context.Context) string {
		return "00000000-0000-0000-0000-000000000000"
	}

	log = logger.NewWithEvents(os.Stdout, logger.LevelInfo, "METRICS", traceIDFn, events)

	// -------------------------------------------------------------------------

	ctx := context.Background()

	if err := run(ctx, log); err != nil {
		log.Error(ctx, "startup", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *logger.Logger) error {

	// -------------------------------------------------------------------------
	// GOMAXPROCS

	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// -------------------------------------------------------------------------
	// Configuration

	cfg := struct {
		conf.Version
		Web struct {
			DebugHost string `conf:"default:0.0.0.0:4010"`
		}
		Expvar struct {
			Host            string        `conf:"default:0.0.0.0:4000"`
			Route           string        `conf:"default:/metrics"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
		Prometheus struct {
			Host            string        `conf:"default:0.0.0.0:4020"`
			Route           string        `conf:"default:/metrics"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
		Collect struct {
			From string `conf:"default:http://localhost:3010/debug/vars"`
		}
		Publish struct {
			To       string        `conf:"default:console"`
			Interval time.Duration `conf:"default:5s"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "copyright information here",
		},
	}

	const prefix = "METRICS"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// -------------------------------------------------------------------------
	// App Starting

	log.Info(ctx, "starting service", "version", build)
	defer log.Info(ctx, "shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Info(ctx, "startup", "config", out)

	log.BuildInfo(ctx)

	// -------------------------------------------------------------------------
	// Start Debug Service

	go func() {
		log.Info(ctx, "startup", "status", "debug router started", "host", cfg.Web.DebugHost)

		if err := http.ListenAndServe(cfg.Web.DebugHost, debug.Mux()); err != nil {
			log.Error(ctx, "shutdown", "status", "debug router closed", "host", cfg.Web.DebugHost, "err", err)
		}
	}()

	// -------------------------------------------------------------------------
	// Start Prometheus Service

	prom := prometheussrv.New(log, cfg.Prometheus.Host, cfg.Prometheus.Route, cfg.Prometheus.ReadTimeout, cfg.Prometheus.WriteTimeout, cfg.Prometheus.IdleTimeout)
	defer prom.Stop(cfg.Prometheus.ShutdownTimeout)

	// -------------------------------------------------------------------------
	// Start expvar Service

	exp := expvarsrv.New(log, cfg.Expvar.Host, cfg.Expvar.Route, cfg.Expvar.ReadTimeout, cfg.Expvar.WriteTimeout, cfg.Expvar.IdleTimeout)
	defer exp.Stop(cfg.Expvar.ShutdownTimeout)

	// -------------------------------------------------------------------------
	// Start collectors and publishers

	collector, err := collector.New(cfg.Collect.From)
	if err != nil {
		return fmt.Errorf("starting collector: %w", err)
	}

	stdout := publisher.NewStdout(log)

	publish, err := publisher.New(log, collector, cfg.Publish.Interval, prom.Publish, exp.Publish, stdout.Publish)
	if err != nil {
		return fmt.Errorf("starting publisher: %w", err)
	}
	defer publish.Stop()

	// -------------------------------------------------------------------------
	// Shutdown

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	log.Info(ctx, "shutdown", "status", "shutdown started")
	defer log.Info(ctx, "shutdown", "status", "shutdown complete")

	return nil
}
