package main

import (
	"context"
	"expvar" // Calls init function.
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ardanlabs/conf"
	"github.com/ardanlabs/service/app/sales-api/handlers"
	"github.com/ardanlabs/service/business/sys/auth"
	"github.com/ardanlabs/service/business/sys/metrics"
	"github.com/ardanlabs/service/foundation/database"
	"github.com/ardanlabs/service/foundation/keystore"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

/*
Need to figure out timeouts for http service.
*/

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {

	// Construct the application logger.
	log := logger.New("SALES-API")
	defer log.Sync()

	// Make sure the program is using the correct
	// number of threads if a CPU quota is set.
	if _, err := maxprocs.Set(); err != nil {
		log.Errorw("startup", zap.Error(err))
		os.Exit(1)
	}
	log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// Perform the startup and shutdown sequence.
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
			APIHost         string        `conf:"default:0.0.0.0:3000"`
			DebugHost       string        `conf:"default:0.0.0.0:4000"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:20s"`
		}
		Auth struct {
			KeysFolder string `conf:"default:zarf/keys/"`
			Algorithm  string `conf:"default:RS256"`
		}
		DB struct {
			User         string `conf:"default:postgres"`
			Password     string `conf:"default:postgres,mask"`
			Host         string `conf:"default:db"`
			Name         string `conf:"default:postgres"`
			MaxIdleConns int    `conf:"default:0"`
			MaxOpenConns int    `conf:"default:0"`
			DisableTLS   bool   `conf:"default:true"`
		}
		Zipkin struct {
			ReporterURI string  `conf:"default:http://localhost:9411/api/v2/spans"`
			ServiceName string  `conf:"default:sales-api"`
			Probability float64 `conf:"default:0.05"`
		}
	}{
		Version: conf.Version{
			SVN:  build,
			Desc: "copyright information here",
		},
	}

	if err := conf.Parse(os.Args[1:], "SALES", &cfg); err != nil {
		switch err {
		case conf.ErrHelpWanted:
			usage, err := conf.Usage("SALES", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}
			fmt.Println(usage)
			return nil
		case conf.ErrVersionWanted:
			version, err := conf.VersionString("SALES", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config version")
			}
			fmt.Println(version)
			return nil
		}
		return errors.Wrap(err, "parsing config")
	}

	// =========================================================================
	// App Starting

	expvar.NewString("build").Set(build)
	log.Infow("starting service", "version", build)
	defer log.Infow("shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return errors.Wrap(err, "generating config for output")
	}
	log.Infow("startup", "config", out)

	// =========================================================================
	// Initialize authentication support

	log.Infow("startup", "status", "initializing authentication support")

	// Construct a key store based on the key files stored in
	// the specified directory.
	ks, err := keystore.NewFS(os.DirFS(cfg.Auth.KeysFolder))
	if err != nil {
		return errors.Wrap(err, "reading keys")
	}

	auth, err := auth.New(cfg.Auth.Algorithm, ks)
	if err != nil {
		return errors.Wrap(err, "constructing auth")
	}

	// =========================================================================
	// Start Database

	log.Infow("startup", "status", "initializing database support", "host", cfg.DB.Host)

	db, err := database.Open(database.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	})
	if err != nil {
		return errors.Wrap(err, "connecting to db")
	}
	defer func() {
		log.Infow("shutdown", "status", "stopping database support", "host", cfg.DB.Host)
		db.Close()
	}()

	// =========================================================================
	// Start Tracing Support

	// WARNING: The current Init settings are using defaults which may not be
	// compatible with your project. Please review the documentation for
	// opentelemetry.

	log.Infow("startup", "status", "initializing OT/Zipkin tracing support")

	exporter, err := zipkin.New(
		cfg.Zipkin.ReporterURI,
		// zipkin.WithLogger(zap.NewStdLog(log)),
	)
	if err != nil {
		return errors.Wrap(err, "creating new exporter")
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.TraceIDRatioBased(cfg.Zipkin.Probability)),
		trace.WithBatcher(exporter,
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
			trace.WithBatchTimeout(trace.DefaultBatchTimeout),
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
		),
		trace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(cfg.Zipkin.ServiceName),
				attribute.String("exporter", "zipkin"),
			),
		),
	)

	// I can only get this working properly using the singleton :(
	otel.SetTracerProvider(traceProvider)
	defer traceProvider.Shutdown(context.Background())

	// =========================================================================
	// Start Debug Service

	log.Infow("startup", "status", "debug router started", "host", cfg.Web.DebugHost)

	// The Debug function returns a mux to listen and serve on for all the debug
	// related endpoints. This include the standard library endpoints.

	// Construct the mux for the debug calls.
	debugMux := handlers.DebugMux(build, log, db)

	// Start the service listening for debug requests.
	// Not concerned with shutting this down with load shedding.
	go func() {
		if err := http.ListenAndServe(cfg.Web.DebugHost, debugMux); err != nil {
			log.Errorw("shutdown", "status", "debug router closed", "host", cfg.Web.DebugHost, "ERROR", err)
		}
	}()

	// =========================================================================
	// Start API Service

	log.Infow("startup", "status", "initializing API support")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Construct the mux for the API calls.
	apiMux := handlers.APIMux(handlers.APIMuxConfig{
		Shutdown: shutdown,
		Log:      log,
		Metrics:  metrics.New(),
		Auth:     auth,
		DB:       db,
	})

	// Construct a server to service the requests against the mux.
	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      apiMux,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     zap.NewStdLog(log.Desugar()),
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for api requests.
	go func() {
		log.Infow("startup", "status", "api router started", "host", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// =========================================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return errors.Wrap(err, "server error")

	case sig := <-shutdown:
		log.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", "shutdown complete", "signal", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and shed load.
		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return errors.Wrap(err, "could not stop server gracefully")
		}
	}

	return nil
}
