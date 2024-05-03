package mid

import (
	"context"

	"github.com/ardanlabs/service/app/api/metrics"
)

// Metrics updates program counters.
func Metrics(ctx context.Context, handler Handler) (any, error) {
	ctx = metrics.Set(ctx)

	resp, err := handler(ctx)

	n := metrics.AddRequests(ctx)

	if n%1000 == 0 {
		metrics.AddGoroutines(ctx)
	}

	if err != nil {
		metrics.AddErrors(ctx)
	}

	return resp, err
}
