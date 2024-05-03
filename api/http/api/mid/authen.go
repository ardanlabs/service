package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/api/auth"
	"github.com/ardanlabs/service/app/api/authclient"
	"github.com/ardanlabs/service/app/api/mid"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// Authenticate validates authentication via the auth service.
func Authenticate(log *logger.Logger, client *authclient.Client) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
			hdl := func(ctx context.Context) (any, error) {
				return handler(ctx, w, r)
			}

			return mid.Authenticate(ctx, log, client, r.Header.Get("authorization"), hdl)
		}

		return h
	}

	return m
}

// Bearer processes JWT authentication logic.
func Bearer(ath *auth.Auth) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
			hdl := func(ctx context.Context) (any, error) {
				return handler(ctx, w, r)
			}

			return mid.Bearer(ctx, ath, r.Header.Get("authorization"), hdl)
		}

		return h
	}

	return m
}

// Basic processes basic authentication logic.
func Basic(userBus *userbus.Business, ath *auth.Auth) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
			hdl := func(ctx context.Context) (any, error) {
				return handler(ctx, w, r)
			}

			return mid.Basic(ctx, ath, userBus, r.Header.Get("authorization"), hdl)
		}

		return h
	}

	return m
}
