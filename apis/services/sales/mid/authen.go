package mid

import (
	"context"
	"net/http"
	"time"

	"github.com/ardanlabs/service/app/api/authsrv"
	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/app/api/mid"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// Authenticate validates a JWT from the `Authorization` header.
func Authenticate(log *logger.Logger, authSrv *authsrv.AuthSrv) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()

			resp, err := authSrv.Authenticate(ctx, r.Header.Get("authorization"))
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			ctx = mid.SetUserID(ctx, resp.UserID)
			ctx = mid.SetClaims(ctx, resp.Claims)

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
