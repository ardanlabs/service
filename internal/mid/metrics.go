package mid

import (
	"context"
	"expvar"
	"net/http"
	"runtime"

	"github.com/ardanlabs/service/internal/platform/web"
	"go.opencensus.io/trace"
)

// m contains the global program counters for the application.
var m = struct {
	gr  *expvar.Int
	req *expvar.Int
	err *expvar.Int
}{
	gr:  expvar.NewInt("goroutines"),
	req: expvar.NewInt("requests"),
	err: expvar.NewInt("errors"),
}

// Metrics writes some information about the request to the logs in the
func Metrics() web.Middleware {

	// This is the actual middleware function to be execute.
	f := func(before web.Handler) web.Handler {

		// Wrap this handler around the next one provided.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
			ctx, span := trace.StartSpan(ctx, "internal.mid.Metrics")
			defer span.End()

			err := before(ctx, w, r, params)

			// Add one to the request counter.
			m.req.Add(1)

			// Include the current count for the number of goroutines.
			if m.req.Value()%100 == 0 {
				m.gr.Set(int64(runtime.NumGoroutine()))
			}

			// Add one to the errors counter if an error occurred
			// on this request.
			if err != nil {
				m.err.Add(1)
			}

			// Return the error so it can be handled further up the chain.
			return err
		}

		return h
	}

	return f
}
