// Package mid contains the set of values the middleware handlers for using
// the http protocol.
package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/sdk/mid"
	"github.com/ardanlabs/service/foundation/web"
)

type midFunc func(ctx context.Context, r *http.Request, next mid.HandlerFunc) (mid.Encoder, error)

func addMidFunc(midFunc midFunc) web.MidFunc {
	m := func(webHandlerFunc web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) (web.Encoder, error) {
			next := func(ctx context.Context) (mid.Encoder, error) {
				return webHandlerFunc(ctx, r)
			}

			return midFunc(ctx, r, next)
		}

		return h
	}

	return m
}
