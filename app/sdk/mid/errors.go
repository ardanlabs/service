package mid

import (
	"context"
	"path"

	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/tracer"
)

// Errors handles errors coming out of the call chain.
func Errors(ctx context.Context, log *logger.Logger, next HandlerFunc) Encoder {
	resp := next(ctx)
	err := isError(resp)
	if err == nil {
		return resp
	}

	_, span := tracer.AddSpan(ctx, "app.sdk.mid.error")
	span.RecordError(err)
	defer span.End()

	appErr, ok := err.(*errs.Error)
	if !ok {
		appErr = errs.Newf(errs.Internal, "Internal Server Error")
	}

	log.Error(ctx, "handled error during request", "err", err, "source_err_file", path.Base(appErr.FileName), "source_err_func", path.Base(appErr.FuncName))

	// Send the error to the transport package so the error can be
	// used as the response.

	return appErr
}
