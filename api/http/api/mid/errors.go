package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/api/mid"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// Errors executes the errors middleware functionality.
func Errors(log *logger.Logger) web.Middleware {
	midFunc := func(ctx context.Context, r *http.Request, next mid.APIHandler) (mid.Encoder, error) {
		return mid.Errors(ctx, log, next)
	}

	return addMiddleware(midFunc)
}
