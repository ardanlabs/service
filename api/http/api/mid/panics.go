package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/api/http/api/response"
	"github.com/ardanlabs/service/app/api/mid"
	"github.com/ardanlabs/service/foundation/web"
)

// Panics executes the panic middleware functionality.
func Panics() web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
			hdl := func(ctx context.Context) mid.Response {
				return response.ToMid(handler(ctx, w, r))
			}

			return response.ToWeb(mid.Panics(ctx, hdl))
		}

		return h
	}

	return m
}
