package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/api/mid"
	"github.com/ardanlabs/service/foundation/web"
)

// Metrics updates program counters using the middleware functionality.
func Metrics() web.MidHandler {
	midFunc := func(ctx context.Context, w http.ResponseWriter, r *http.Request, hdl mid.Handler) (any, error) {
		return mid.Metrics(ctx, hdl)
	}

	return addMiddleware(midFunc)
}
