package mid

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/foundation/logger"
)

// Logger writes information about the request to the logs.
func Logger(ctx context.Context, log *logger.Logger, path string, rawQuery string, method string, remoteAddr string, next HandlerFunc) Encoder {
	now := time.Now()

	if rawQuery != "" {
		path = fmt.Sprintf("%s?%s", path, rawQuery)
	}

	log.Info(ctx, "request started", "method", method, "path", path, "remoteaddr", remoteAddr)

	resp := next(ctx)
	err := isError(resp)

	var statusCode = errs.OK
	if err != nil {
		statusCode = errs.Internal

		var v *errs.Error
		if errors.As(err, &v) {
			statusCode = v.Code
		}
	}

	log.Info(ctx, "request completed", "method", method, "path", path, "remoteaddr", remoteAddr,
		"statuscode", statusCode, "since", time.Since(now).String())

	return resp
}
