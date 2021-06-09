package mid

import (
	"context"
	"net/http"
	"runtime"
	"strings"

	"github.com/ardanlabs/service/business/sys/metrics"
	"github.com/ardanlabs/service/foundation/web"
	"go.opentelemetry.io/otel/trace"
)

// =============================================================================

// Metrics updates program counters.
func Metrics(data *metrics.Metrics) web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.web.mid.metrics")
			defer span.End()

			// Add the metrics value for metric gathering.
			ctx = context.WithValue(ctx, metrics.Key, data)

			// Call the next handler.
			err := handler(ctx, w, r)

			// Handle updating the metrics that can be handled here.

			// Don't count anything on /debug routes towards metrics.
			// Call the next handler to continue processing.
			if strings.HasPrefix(r.URL.Path, "/debug") {
				return err
			}

			// Increment the request counter.
			data.Requets.Add(1)

			// Increment if there is an error flowing through the request.
			if err != nil {
				data.Errors.Add(1)
			}

			// Update the count for the number of active goroutines every 100 requests.
			if data.Requets.Value()%100 == 0 {
				data.Goroutines.Set(int64(runtime.NumGoroutine()))
			}

			// Return the error so it can be handled further up the chain.
			return err
		}

		return h
	}

	return m
}
