package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/api/mid"
	"github.com/ardanlabs/service/foundation/web"
)

// Metrics updates program counters using the middleware functionality.
func Metrics() web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			f := func(ctx context.Context) error {
				return handler(ctx, w, r)
			}

			return mid.Metrics(ctx, f)
		}

		return h
	}

	return m
}
