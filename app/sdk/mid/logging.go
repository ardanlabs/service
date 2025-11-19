package mid

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// Logger writes information about the request to the logs.
func Logger(log *logger.Logger) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			now := time.Now()

			path := r.URL.Path
			if r.URL.RawQuery != "" {
				path = fmt.Sprintf("%s?%s", path, r.URL.RawQuery)
			}

			log.Info(ctx, "request started", "method", r.Method, "path", path, "remoteaddr", r.RemoteAddr)

			resp := next(ctx, r)

			var statusCode = errs.None
			if err := checkIsError(resp); err != nil {
				statusCode = errs.Internal

				var appErr *errs.Error
				if errors.As(err, &appErr) {
					statusCode = appErr.Code
				}
			}

			log.Info(ctx, "request completed", "method", r.Method, "path", path, "remoteaddr", r.RemoteAddr,
				"statuscode", statusCode, "since", time.Since(now).String())

			return resp
		}

		return h
	}

	return m
}
