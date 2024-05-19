package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/sdk/mid"
	"github.com/ardanlabs/service/foundation/web"
)

// Metrics updates program counters using the middleware functionality.
func Metrics() web.MidFunc {
	midFunc := func(ctx context.Context, r *http.Request, next mid.HandlerFunc) (mid.Encoder, error) {
		return mid.Metrics(ctx, next)
	}

	return addMidFunc(midFunc)
}
