package trace

import (
	"context"
	"log"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/trace/zipkin"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// InitTracer creates a new trace provider instance and registers it as global trace provider.
func InitTracer(ServiceName string, LocalEndpoint string, ReporterURI string, logger *log.Logger) error {
	// Create Zipkin Exporter
	exporter, err := zipkin.NewExporter(
		ReporterURI,
		ServiceName,
		zipkin.WithLogger(logger),
	)
	if err != nil {
		return err
	}

	// For demoing purposes, always sample. In a production application, you should
	// configure this to a trace.ProbabilitySampler set at the desired
	// probability.
	tp, err := sdktrace.NewProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5),
			sdktrace.WithMaxExportBatchSize(10),
		),
	)
	if err != nil {
		return err
	}
	global.SetTraceProvider(tp)
	return nil
}

// NewSpan start a span with trace
func NewSpan(ctx context.Context, name string) context.Context {
	ctx, span := global.Tracer("main").Start(ctx, name)
	defer span.End()
	return ctx
}
