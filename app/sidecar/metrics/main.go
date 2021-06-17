package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ardanlabs/conf"
	"github.com/ardanlabs/service/app/sales-api/handlers"
	"github.com/ardanlabs/service/app/sidecar/metrics/collector"
	"github.com/ardanlabs/service/app/sidecar/metrics/publisher"
	"github.com/ardanlabs/service/app/sidecar/metrics/publisher/expvar"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/pkg/errors"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {

	// Construct the application logger.
	log := logger.New("METRICS")
	defer log.Sync()

	// Make sure the program is using the correct
	// number of threads if a CPU quota is set.
	if _, err := maxprocs.Set(); err != nil {
		log.Errorw("startup", zap.Error(err))
		os.Exit(1)
	}
	log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	if err := run(log); err != nil {
		log.Errorw("startup", "ERROR", err)
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {

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
			SVN:  build,
			Desc: "copyright information here",
		},
	}

	if err := conf.Parse(os.Args[1:], "METRICS", &cfg); err != nil {
		switch err {
		case conf.ErrHelpWanted:
			usage, err := conf.Usage("METRICS", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}
			fmt.Println(usage)
			return nil
		case conf.ErrVersionWanted:
			version, err := conf.VersionString("METRICS", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config version")
			}
			fmt.Println(version)
			return nil
		}
		return errors.Wrap(err, "parsing config")
	}

	out, err := conf.String(&cfg)
	if err != nil {
		return errors.Wrap(err, "generating config for output")
	}
	log.Infow("startup", "config", out)

	// =========================================================================
	// Start Debug Service

	log.Infow("startup", "status", "debug router started", "host", cfg.Web.DebugHost)

	// The Debug function returns a mux to listen and serve on for all the debug
	// related endpoints. This include the standard library endpoints.

	debugMux := handlers.DebugStandardLibraryMux()

	// Start the service listening for debug requests.
	// Not concerned with shutting this down with load shedding.
	go func() {
		if err := http.ListenAndServe(cfg.Web.DebugHost, debugMux); err != nil {
			log.Errorw("shutdown", "status", "debug router closed", "host", cfg.Web.DebugHost, "ERROR", err)
		}
	}()

	// =========================================================================
	// Start expvar Service

	exp := expvar.New(log, cfg.Expvar.Host, cfg.Expvar.Route, cfg.Expvar.ReadTimeout, cfg.Expvar.WriteTimeout, cfg.Expvar.IdleTimeout)
	defer exp.Stop(cfg.Expvar.ShutdownTimeout)

	// =========================================================================
	// Start collectors and publishers

	// Initialize to allow for the collection of metrics.
	collector, err := collector.New(cfg.Collect.From)
	if err != nil {
		return errors.Wrap(err, "starting collector")
	}

	// Create a stdout publisher.
	// TODO: Respect the cfg.publish.to config option.
	stdout := publisher.NewStdout(log)

	// Start the publisher to collect/publish metrics.
	publish, err := publisher.New(log, collector, cfg.Publish.Interval, exp.Publish, stdout.Publish)
	if err != nil {
		return errors.Wrap(err, "starting publisher")
	}
	defer publish.Stop()

	// =========================================================================
	// Shutdown

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	log.Infow("shutdown", "status", "shutdown started")
	defer log.Infow("shutdown", "status", "shutdown complete")

	return nil
}
