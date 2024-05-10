package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/sdk/auth"
	"github.com/ardanlabs/service/app/sdk/authclient"
	"github.com/ardanlabs/service/app/sdk/mid"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// Authenticate validates authentication via the auth service.
func Authenticate(log *logger.Logger, client *authclient.Client) web.Middleware {
	midFunc := func(ctx context.Context, r *http.Request, next mid.Handler) (mid.Encoder, error) {
		return mid.Authenticate(ctx, log, client, r.Header.Get("authorization"), next)
	}

	return addMiddleware(midFunc)
}

// Bearer processes JWT authentication logic.
func Bearer(ath *auth.Auth) web.Middleware {
	midFunc := func(ctx context.Context, r *http.Request, next mid.Handler) (mid.Encoder, error) {
		return mid.Bearer(ctx, ath, r.Header.Get("authorization"), next)
	}

	return addMiddleware(midFunc)
}

// Basic processes basic authentication logic.
func Basic(userBus *userbus.Business, ath *auth.Auth) web.Middleware {
	midFunc := func(ctx context.Context, r *http.Request, next mid.Handler) (mid.Encoder, error) {
		return mid.Basic(ctx, ath, userBus, r.Header.Get("authorization"), next)
	}

	return addMiddleware(midFunc)
}
