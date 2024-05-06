package mid

import (
	"context"
	"path"

	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/tracer"
)

// Errors handles errors coming out of the call chain. It detects normal
// application errors which are used to respond to the client in a uniform way.
// Unexpected errors (status >= 500) are logged.
func Errors(ctx context.Context, log *logger.Logger, next Handler) (Encoder, error) {
	resp, err := next(ctx)
	if err == nil {
		return resp, nil
	}

	switch v := err.(type) {
	case *errs.Error:
		log.Error(ctx, "message", "ERROR", err, "FileName", path.Base(v.FileName), "FuncName", path.Base(v.FuncName))

	default:
		log.Error(ctx, "message", "ERROR", err)
	}

	_, span := tracer.AddSpan(ctx, "app.api.mid.error")
	span.RecordError(err)
	defer span.End()

	// Send the error to the web package so the error can be
	// used as the response.

	return nil, err
}
