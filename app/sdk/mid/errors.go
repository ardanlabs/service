package mid

import (
	"context"
	"path"

	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/tracer"
)

// Errors handles errors coming out of the call chain.
func Errors(ctx context.Context, log *logger.Logger, next Handler) (Encoder, error) {
	resp, err := next(ctx)
	if err == nil {
		return resp, nil
	}

	_, span := tracer.AddSpan(ctx, "app.api.mid.error")
	span.RecordError(err)
	defer span.End()

	v, ok := err.(*errs.Error)
	if !ok {
		v = errs.New(errs.Internal, err)
		err = v
	}

	log.Error(ctx, "message", "ERROR", err, "FileName", path.Base(v.FileName), "FuncName", path.Base(v.FuncName))

	// Send the error to the transport package so the error can be
	// used as the response.

	return nil, err
}
