package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/foundation/web"
	"go.opentelemetry.io/otel/trace"
)

// Cors sets the response headers needed for Cross-Origin Resource Sharing
func Cors(origin string) web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.mid.cors")
			defer span.End()

			// Set CORS Headers
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

			// Call the next handler.
			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
