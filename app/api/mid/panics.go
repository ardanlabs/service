package mid

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/ardanlabs/service/app/api/metrics"
)

// Panics recovers from panics and converts the panic to an error so it is
// reported in Metrics and handled in Errors.
func Panics(ctx context.Context, handler func(context.Context) error) (err error) {

	// Defer a function to recover from a panic and set the err return
	// variable after the fact.
	defer func() {
		if rec := recover(); rec != nil {
			trace := debug.Stack()
			err = fmt.Errorf("PANIC [%v] TRACE[%s]", rec, string(trace))

			metrics.AddPanics(ctx)
		}
	}()

	return handler(ctx)
}
