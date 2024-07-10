package mid

import (
	"context"

	"github.com/ardanlabs/service/app/sdk/metrics"
)

// Metrics updates program counters.
func Metrics(ctx context.Context, next HandlerFunc) Encoder {
	ctx = metrics.Set(ctx)

	resp := next(ctx)

	n := metrics.AddRequests(ctx)

	if n%1000 == 0 {
		metrics.AddGoroutines(ctx)
	}

	if isError(resp) != nil {
		metrics.AddErrors(ctx)
	}

	return resp
}
