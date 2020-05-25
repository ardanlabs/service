package tracer

import (
	"log"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/trace/zipkin"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Init creates a new trace provider instance and registers it as global trace provider.
func Init(ServiceName string, LocalEndpoint string, ReporterURI string, log *log.Logger) error {
	exporter, err := zipkin.NewExporter(
		ReporterURI,
		ServiceName,
		zipkin.WithLogger(log),
	)
	if err != nil {
		return errors.Wrap(err, "creating new exporter")
	}

	// For demoing purposes, always sample. In a production application, you should
	// configure this to a trace.ProbabilitySampler set at the desired probability.
	tp, err := sdktrace.NewProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5),
			sdktrace.WithMaxExportBatchSize(10),
		),
	)
	if err != nil {
		return errors.Wrap(err, "creating new provider")
	}

	global.SetTraceProvider(tp)
	return nil
}
