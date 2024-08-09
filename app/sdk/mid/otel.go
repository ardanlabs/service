package mid

import (
	"context"

	"github.com/ardanlabs/service/foundation/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Otel starts the otel tracing and stores the trace id in the context.
func Otel(ctx context.Context, tracer trace.Tracer, endpoint string, headerCarrier propagation.HeaderCarrier, next HandlerFunc) Encoder {
	ctx, span := otel.StartTrace(ctx, tracer, "start.request", endpoint, headerCarrier)
	defer span.End()

	return next(ctx)
}
