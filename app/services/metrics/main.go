package main

import (
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/ardanlabs/service/app/services/metrics/collector"
	"github.com/ardanlabs/service/app/services/metrics/publisher"
	expvarsrv "github.com/ardanlabs/service/app/services/metrics/publisher/expvar"
	"github.com/ardanlabs/service/foundation/logger"
	"go.uber.org/zap"
)

var build = "develop"

func main() {
	log, err := logger.New("METRICS")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync()

	if err := run(log); err != nil {
		log.Errorw("startup", "ERROR", err)
		log.Sync()
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {

	// =========================================================================
	// GOMAXPROCS

	log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// =========================================================================
	// Configuration

	cfg := struct {
		conf.Version
		Web struct {
			DebugHost string `conf:"default:0.0.0.0:4001"`
		}
		Expvar struct {
			Host            string        `conf:"default:0.0.0.0:3001"`
			Route           string        `conf:"default:/metrics"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
		Collect struct {
			From string `conf:"default:http://localhost:4000/debug/vars"`
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

	// =========================================================================
	// App Starting

	log.Infow("starting service", "version", build)
	defer log.Infow("shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Infow("startup", "config", out)

	// =========================================================================
	// Start Debug Service

	log.Infow("startup", "status", "debug router started", "host", cfg.Web.DebugHost)

	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars", expvar.Handler())

	go func() {
		if err := http.ListenAndServe(cfg.Web.DebugHost, mux); err != nil {
			log.Errorw("shutdown", "status", "debug router closed", "host", cfg.Web.DebugHost, "ERROR", err)
		}
	}()

	// =========================================================================
	// Start expvar Service

	exp := expvarsrv.New(log, cfg.Expvar.Host, cfg.Expvar.Route, cfg.Expvar.ReadTimeout, cfg.Expvar.WriteTimeout, cfg.Expvar.IdleTimeout)
	defer exp.Stop(cfg.Expvar.ShutdownTimeout)

	// =========================================================================
	// Start collectors and publishers

	collector, err := collector.New(cfg.Collect.From)
	if err != nil {
		return fmt.Errorf("starting collector: %w", err)
	}

	stdout := publisher.NewStdout(log)

	publish, err := publisher.New(log, collector, cfg.Publish.Interval, exp.Publish, stdout.Publish)
	if err != nil {
		return fmt.Errorf("starting publisher: %w", err)
	}
	defer publish.Stop()

	// =========================================================================
	// Shutdown

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	log.Infow("shutdown", "status", "shutdown started")
	defer log.Infow("shutdown", "status", "shutdown complete")

	return nil
}
