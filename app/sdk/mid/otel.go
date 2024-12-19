package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/foundation/otel"
	"github.com/ardanlabs/service/foundation/web"
	"go.opentelemetry.io/otel/trace"
)

// Otel starts the otel tracing and stores the trace id in the context.
func Otel(tracer trace.Tracer) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			ctx = otel.InjectTracing(ctx, tracer)

			return next(ctx, r)
		}

		return h
	}

	return m
}
