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
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			err := handler(ctx, w, r)
			if err == nil {
				return nil
			}

			ctx, errs := mid.Errors(ctx, log, err)

			if err := web.Respond(ctx, w, errs, errs.Details.HTTPStatusCode); err != nil {
				return err
			}

			// If we receive the shutdown err we need to return it
			// back to the base handler to shut down the service.
			if web.IsShutdown(err) {
				return err
			}

			return nil
		}

		return h
	}

	return m
}
