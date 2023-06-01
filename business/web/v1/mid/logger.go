package mid

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ardanlabs/service/foundation/web"
	"golang.org/x/exp/slog"
)

// Logger writes information about the request to the logs.
func Logger(log *slog.Logger) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			v := web.GetValues(ctx)

			path := r.URL.Path
			if r.URL.RawQuery != "" {
				path = fmt.Sprintf("%s?%s", path, r.URL.RawQuery)
			}

			log.Info("request started", "trace_id", v.TraceID, "method", r.Method, "path", path,
				"remoteaddr", r.RemoteAddr)

			err := handler(ctx, w, r)

			log.Info("request completed", "trace_id", v.TraceID, "method", r.Method, "path", path,
				"remoteaddr", r.RemoteAddr, "statuscode", v.StatusCode, "since", time.Since(v.Now))

			return err
		}

		return h
	}

	return m
}
