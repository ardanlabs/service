package mid

import (
	"context"

	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/tracer"
)

// Errors handles errors coming out of the call chain. It detects normal
// application errors which are used to respond to the client in a uniform way.
// Unexpected errors (status >= 500) are logged.
func Errors(ctx context.Context, log *logger.Logger, handler Handler) Response {
	resp := handler(ctx)
	if resp.Errs == nil {
		return resp
	}

	log.Error(ctx, "message", "ERROR", resp.Errs)

	_, span := tracer.AddSpan(ctx, "app.api.mid.error")
	span.RecordError(resp.Errs)
	defer span.End()

	return resp
}
