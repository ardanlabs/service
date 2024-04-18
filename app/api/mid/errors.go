package mid

import (
	"context"

	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// Errors handles errors coming out of the call chain. It detects normal
// application errors which are used to respond to the client in a uniform way.
// Unexpected errors (status >= 500) are logged.
func Errors(ctx context.Context, log *logger.Logger, err error) (context.Context, errs.Error) {
	log.Error(ctx, "message", "ERROR", err.Error())

	ctx, span := web.AddSpan(ctx, "app.api.mid.http.error")
	span.RecordError(err)
	defer span.End()

	if errs.IsError(err) {
		return ctx, errs.GetError(err)
	}

	return ctx, errs.Newf(errs.Unknown, errs.Unknown.String())
}
