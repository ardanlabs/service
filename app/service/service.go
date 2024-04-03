// Package service initialized the service for http usage.
package service

import (
	"context"
	"expvar"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/core/crud/delegate"
	"github.com/ardanlabs/service/business/core/crud/home"
	"github.com/ardanlabs/service/business/core/crud/home/stores/homedb"
	"github.com/ardanlabs/service/business/core/crud/product"
	"github.com/ardanlabs/service/business/core/crud/product/stores/productdb"
	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/core/crud/user/stores/userdb"
	"github.com/ardanlabs/service/business/core/views/vproduct"
	"github.com/ardanlabs/service/business/core/views/vproduct/stores/vproductdb"
	"github.com/ardanlabs/service/business/data/sqldb"
	"github.com/ardanlabs/service/foundation/keystore"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	otrace "go.opentelemetry.io/otel/trace"
)

// Service represents all the things the service needs to run.
type Service struct {
	Log      *logger.Logger
	DB       *sqlx.DB
	Auth     *auth.Auth
	Tracer   otrace.Tracer
	Provider *trace.TracerProvider
	BusCrud  BusCrud
	BusView  BusView
}

// New is called to create a new encore Service.
func New(ctx context.Context, log *logger.Logger, cfg Config) (*Service, error) {

	// -------------------------------------------------------------------------
	// GOMAXPROCS

	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// -------------------------------------------------------------------------
	// App Starting

	log.Info(ctx, "starting service", "version", cfg.Version.Build)
	defer log.Info(ctx, "shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return nil, fmt.Errorf("generating config for output: %w", err)
	}
	log.Info(ctx, "startup", "config", out)

	expvar.NewString("build").Set(cfg.Version.Build)

	// -------------------------------------------------------------------------
	// Database Support

	log.Info(ctx, "startup", "status", "initializing database support", "hostport", cfg.DB.HostPort)

	db, err := sqldb.Open(sqldb.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		HostPort:     cfg.DB.HostPort,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	})
	if err != nil {
		return nil, fmt.Errorf("connecting to db: %w", err)
	}

	// -------------------------------------------------------------------------
	// Initialize authentication support

	log.Info(ctx, "startup", "status", "initializing authentication support")

	// Load the private keys files from disk. We can assume some system like
	// Vault has created these files already. How that happens is not our
	// concern.
	ks := keystore.New()
	if err := ks.LoadRSAKeys(os.DirFS(cfg.Auth.KeysFolder)); err != nil {
		return nil, fmt.Errorf("reading keys: %w", err)
	}

	authCfg := auth.Config{
		Log:       log,
		DB:        db,
		KeyLookup: ks,
	}

	auth, err := auth.New(authCfg)
	if err != nil {
		return nil, fmt.Errorf("constructing auth: %w", err)
	}

	// -------------------------------------------------------------------------
	// Start Tracing Support

	log.Info(ctx, "startup", "status", "initializing tracing support")

	traceProvider, err := startTracing(
		cfg.Tempo.ServiceName,
		cfg.Tempo.ReporterURI,
		cfg.Tempo.Probability,
	)
	if err != nil {
		return nil, fmt.Errorf("starting tracing: %w", err)
	}

	tracer := traceProvider.Tracer("service")

	// -------------------------------------------------------------------------
	// Build Core APIs

	log.Info(ctx, "startup", "status", "initializing business support")

	delegate := delegate.New(log)
	userCore := user.NewCore(log, delegate, userdb.NewStore(log, db))
	productCore := product.NewCore(log, userCore, delegate, productdb.NewStore(log, db))
	homeCore := home.NewCore(log, userCore, delegate, homedb.NewStore(log, db))
	vproductCore := vproduct.NewCore(vproductdb.NewStore(log, db))

	// -------------------------------------------------------------------------
	// Construct Service value.

	s := Service{
		Log:      log,
		DB:       db,
		Auth:     auth,
		Tracer:   tracer,
		Provider: traceProvider,
		BusCrud: BusCrud{
			Delegate: delegate,
			Home:     homeCore,
			Product:  productCore,
			User:     userCore,
		},
		BusView: BusView{
			Product: vproductCore,
		},
	}

	return &s, nil
}

// Shutdown implements a function that will be called by encore when the service
// is signaled to shutdown.
func (s *Service) Shutdown(ctx context.Context) {
	defer s.Log.Info(ctx, "shutdown", "status", "shutdown complete")

	s.Log.Info(ctx, "shutdown", "status", "stopping tracing provideer")
	s.Provider.Shutdown(context.Background())

	s.Log.Info(ctx, "shutdown", "status", "stopping database support")
	s.DB.Close()
}

// startTracing configure open telemetry to be used with Grafana Tempo.
func startTracing(serviceName string, reporterURI string, probability float64) (*trace.TracerProvider, error) {

	// WARNING: The current settings are using defaults which may not be
	// compatible with your project. Please review the documentation for
	// opentelemetry.

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(), // This should be configurable
			otlptracegrpc.WithEndpoint(reporterURI),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating new exporter: %w", err)
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.TraceIDRatioBased(probability)),
		trace.WithBatcher(exporter,
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
			trace.WithBatchTimeout(trace.DefaultScheduleDelay*time.Millisecond),
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
		),
		trace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(serviceName),
			),
		),
	)

	// We must set this provider as the global provider for things to work,
	// but we pass this provider around the program where needed to collect
	// our traces.
	otel.SetTracerProvider(traceProvider)

	// Chooses the HTTP header formats we extract incoming trace contexts from,
	// and the headers we set in outgoing requests.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return traceProvider, nil
}
