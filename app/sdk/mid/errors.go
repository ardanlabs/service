package mid

import (
	"context"
	"errors"
	"net/http"
	"path"

	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/otel"
	"github.com/ardanlabs/service/foundation/web"
)

// Errors handles errors coming out of the call chain.
func Errors(log *logger.Logger) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			resp := next(ctx, r)
			err := isError(resp)
			if err == nil {
				return resp
			}

			_, span := otel.AddSpan(ctx, "app.sdk.mid.error")
			span.RecordError(err)
			defer span.End()

			var appErr *errs.Error
			if !errors.As(err, &appErr) {
				appErr = errs.Newf(errs.Internal, "Internal Server Error")
			}

			log.Error(ctx, "handled error during request",
				"err", err,
				"source_err_file", path.Base(appErr.FileName),
				"source_err_func", path.Base(appErr.FuncName))

			if appErr.Code == errs.InternalOnlyLog {
				appErr = errs.Newf(errs.Internal, "Internal Server Error")
			}

			// Send the error to the transport package so the error can be
			// used as the response.

			return appErr
		}

		return h
	}

	return m
}
