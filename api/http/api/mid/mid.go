// Package mid contains the set of values the middleware handlers for using
// the http protocol.
package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/api/mid"
	"github.com/ardanlabs/service/foundation/web"
)

type midFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request, hdl mid.Handler) (any, error)

func addMiddleware(midFunc midFunc) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
			hdl := func(ctx context.Context) (any, error) {
				return handler(ctx, w, r)
			}

			return midFunc(ctx, w, r, hdl)
		}

		return h
	}

	return m
}
