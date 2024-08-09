package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/sdk/mid"
	"github.com/ardanlabs/service/foundation/web"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Otel executes the otel middleware functionality.
func Otel(tracer trace.Tracer) web.MidFunc {
	midFunc := func(ctx context.Context, r *http.Request, next mid.HandlerFunc) mid.Encoder {

		// This will allow us to add the otel trace id to the response header.
		w := web.GetWriter(ctx)
		hc := propagation.HeaderCarrier(w.Header())

		return mid.Otel(ctx, tracer, r.RequestURI, hc, next)
	}

	return addMidFunc(midFunc)
}
