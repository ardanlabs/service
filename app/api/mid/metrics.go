package mid

import (
	"context"

	"github.com/ardanlabs/service/app/api/metrics"
)

// Metrics updates program counters.
func Metrics(ctx context.Context, handler Handler) Response {
	ctx = metrics.Set(ctx)

	resp := handler(ctx)

	n := metrics.AddRequests(ctx)

	if n%1000 == 0 {
		metrics.AddGoroutines(ctx)
	}

	if resp.Errs != nil {
		metrics.AddErrors(ctx)
	}

	return resp
}
