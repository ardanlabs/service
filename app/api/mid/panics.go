package mid

import (
	"context"
	"runtime/debug"

	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/app/api/metrics"
)

// Panics recovers from panics and converts the panic to an error so it is
// reported in Metrics and handled in Errors.
func Panics(ctx context.Context, handler Handler) (resp Response) {

	// Defer a function to recover from a panic and set the err return
	// variable after the fact.
	defer func() {
		if rec := recover(); rec != nil {
			trace := debug.Stack()
			resp = appErrorf(errs.Internal, "PANIC [%v] TRACE[%s]", rec, string(trace))

			metrics.AddPanics(ctx)
		}
	}()

	return handler(ctx)
}
