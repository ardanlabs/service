package mid

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ardanlabs/service/foundation/logger"
)

type httpStatus interface {
	HTTPStatus() int
}

// Logger writes information about the request to the logs.
func Logger(ctx context.Context, log *logger.Logger, path string, rawQuery string, method string, remoteAddr string, handler Handler) (any, error) {
	now := time.Now()

	if rawQuery != "" {
		path = fmt.Sprintf("%s?%s", path, rawQuery)
	}

	log.Info(ctx, "request started", "method", method, "path", path, "remoteaddr", remoteAddr)

	resp, err := handler(ctx)

	var statusCode = http.StatusOK
	v, ok := err.(httpStatus)
	if ok {
		statusCode = v.HTTPStatus()
	}

	log.Info(ctx, "request completed", "method", method, "path", path, "remoteaddr", remoteAddr,
		"statuscode", statusCode, "since", time.Since(now).String())

	return resp, err
}
