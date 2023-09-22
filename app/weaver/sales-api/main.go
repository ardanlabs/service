package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ServiceWeaver/weaver"
	"github.com/ardanlabs/conf/v3"
	v1 "github.com/ardanlabs/service/app/services/sales-api/handlers/v1"
	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/debug"
	"github.com/ardanlabs/service/business/data/dbmigrate"
	db "github.com/ardanlabs/service/business/data/dbsql/pgx"
	"github.com/ardanlabs/service/business/web/auth"
	"github.com/ardanlabs/service/foundation/keystore"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/zarf/keys"
	"go.opentelemetry.io/otel"
)

/*
	TODOs:
	* Add secrets API to Service Weaver and use it.
	* More documentation in the dev.toml file or a link where to go.
	* Break things down by domain.
*/

//go:generate weaver generate

const build = "develop"

func main() {
	if err := weaver.Run(context.Background(), serve); err != nil {
		log.Fatal(err)
	}
}

type server struct {
	weaver.Implements[weaver.Main]
	api, debug weaver.Listener
}

func serve(ctx context.Context, s *server) error {
	log := logger.NewWithHandler(s.Logger(ctx).Handler())

	// -------------------------------------------------------------------------
	// GOMAXPROCS

	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// -------------------------------------------------------------------------
	// Configuration

	cfg := struct {
		conf.Version
		Web struct {
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:20s"`
		}
		DB struct {
			User         string `conf:"default:postgres"`
			Password     string `conf:"default:postgres,mask"`
			Host         string `conf:"default:localhost"`
			Name         string `conf:"default:postgres"`
			MaxIdleConns int    `conf:"default:2"`
			MaxOpenConns int    `conf:"default:0"`
			DisableTLS   bool   `conf:"default:true"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "BILL KENNEDY",
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
	// App Starting

	log.Info(ctx, "starting service", "version", build)
	defer log.Info(ctx, "shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Info(ctx, "startup", "config", out)

	expvar.NewString("build").Set(build)

	// -------------------------------------------------------------------------
	// Database Support

	log.Info(ctx, "startup", "status", "initializing database support", "host", cfg.DB.Host)

	db, err := db.Open(db.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	})
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer func() {
		log.Info(ctx, "shutdown", "status", "stopping database support", "URI", cfg.DB.Host)
		db.Close()
	}()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("testing db %w", err)
	}

	// -------------------------------------------------------------------------
	// Populate the Database, if not already populated.

	log.Info(ctx, "startup", "status", "initializing the database", "URI", cfg.DB.Host)
	if err := dbmigrate.Migrate(ctx, db); err != nil {
		return fmt.Errorf("migrate db: %w", err)
	}

	if err := dbmigrate.Seed(ctx, db); err != nil {
		return fmt.Errorf("seed db: %w", err)
	}

	// -------------------------------------------------------------------------
	// Initialize authentication support

	log.Info(ctx, "startup", "status", "initializing authentication support")

	// Use a simple keystore.
	// TODO(spetrovic): Use the Service Weaver Secrets API once we support it.
	ks, err := keystore.NewFS(keys.DevKeysFS)
	if err != nil {
		return fmt.Errorf("reading keys: %w", err)
	}

	authCfg := auth.Config{
		Log:       log,
		DB:        db,
		KeyLookup: ks,
	}

	auth, err := auth.New(authCfg)
	if err != nil {
		return fmt.Errorf("constructing auth: %w", err)
	}

	// -------------------------------------------------------------------------
	// Start Debug Service

	go func() {
		log.Info(ctx, "startup", "status", "debug v1 router started", "host", s.debug)
		mux := debug.Mux()
		mux.HandleFunc(weaver.HealthzURL, weaver.HealthzHandler)
		if err := http.Serve(s.debug, mux); err != nil {
			log.Error(ctx, "shutdown", "status", "debug v1 router closed", "host", s.debug, "msg", err)
		}
	}()

	// -------------------------------------------------------------------------
	// Start API Service

	log.Info(ctx, "startup", "status", "initializing V1 API support")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	cfgMux := v1.APIMuxConfig{
		UsingWeaver: true,
		Build:       build,
		Shutdown:    shutdown,
		Log:         log,
		Auth:        auth,
		DB:          db,
		Tracer:      otel.Tracer("service"),
	}
	apiMux := v1.APIMux(cfgMux, v1.WithCORS("*"))

	api := http.Server{
		Handler:      weaver.InstrumentHandler("service", apiMux),
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Info(ctx, "startup", "status", "api router started", "host", s.api)

		serverErrors <- api.Serve(s.api)
	}()

	// -------------------------------------------------------------------------
	// Shutdown

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Info(ctx, "shutdown", "status", "shutdown started", "signal", sig)
		defer log.Info(ctx, "shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
