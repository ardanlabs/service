package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

// Panics recovers from panics and converts the panic to an error so it is
// reported in Metrics and handled in Errors.
func Panics() web.Middleware {

	// This is the actual middleware function to be executed.
	f := func(after web.Handler) web.Handler {

		// Wrap this handler around the next one provided.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) (err error) {
			ctx, span := trace.StartSpan(ctx, "internal.mid.Panics")
			defer span.End()

			// Defer a function to recover from a panic and set the err return variable
			// after the fact. Using the errors package will generate a stack trace.
			defer func() {
				if r := recover(); r != nil {
					err = errors.Errorf("panic: %v", r)
				}
			}()

			// Call the next Handler and set its return value in the err variable.
			return after(ctx, w, r, params)
		}

		return h
	}

	return f
}
