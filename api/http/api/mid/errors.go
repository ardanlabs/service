package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/api/mid"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// Errors executes the errors middleware functionality.
func Errors(log *logger.Logger) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
			hdl := func(ctx context.Context) (any, error) {
				return handler(ctx, w, r)
			}

			return mid.Errors(ctx, log, hdl)
		}

		return h
	}

	return m
}
