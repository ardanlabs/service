// Package mid contains the set of values the middleware handlers for using
// the http protocol.
package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/sdk/mid"
	"github.com/ardanlabs/service/foundation/web"
)

type midFunc func(ctx context.Context, r *http.Request, next mid.HandlerFunc) mid.Encoder

func addMidFunc(midFunc midFunc) web.MidFunc {
	m := func(webHandlerFunc web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			next := func(ctx context.Context) mid.Encoder {
				return webHandlerFunc(ctx, r)
			}

			return midFunc(ctx, r, next)
		}

		return h
	}

	return m
}
