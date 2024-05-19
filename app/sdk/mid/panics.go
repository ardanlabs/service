package mid

import (
	"context"
	"runtime/debug"

	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/app/sdk/metrics"
)

// Panics recovers from panics and converts the panic to an error so it is
// reported in Metrics and handled in Errors.
func Panics(ctx context.Context, next HandlerFunc) (resp Encoder, err error) {

	// Defer a function to recover from a panic and set the err return
	// variable after the fact.
	defer func() {
		if rec := recover(); rec != nil {
			trace := debug.Stack()
			err = errs.Newf(errs.Internal, "PANIC [%v] TRACE[%s]", rec, string(trace))

			metrics.AddPanics(ctx)
		}
	}()

	return next(ctx)
}
