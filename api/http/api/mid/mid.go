// Package mid contains the set of values the middleware handlers for using
// the http protocol.
package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/api/mid"
	"github.com/ardanlabs/service/foundation/web"
)

type midFunc func(context.Context, *http.Request, mid.APIHandler) (mid.Encoder, error)

func addMiddleware(midFunc midFunc) web.Middleware {
	m := func(webHandler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (web.Encoder, error) {
			next := func(ctx context.Context) (mid.Encoder, error) {
				return webHandler(ctx, w, r)
			}

			return midFunc(ctx, r, next)
		}

		return h
	}

	return m
}
