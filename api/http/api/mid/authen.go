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
	midFunc := func(ctx context.Context, w http.ResponseWriter, r *http.Request, hdl mid.Handler) (any, error) {
		return mid.Authenticate(ctx, log, client, r.Header.Get("authorization"), hdl)
	}

	return addMiddleware(midFunc)
}

// Bearer processes JWT authentication logic.
func Bearer(ath *auth.Auth) web.MidHandler {
	midFunc := func(ctx context.Context, w http.ResponseWriter, r *http.Request, hdl mid.Handler) (any, error) {
		return mid.Bearer(ctx, ath, r.Header.Get("authorization"), hdl)
	}

	return addMiddleware(midFunc)
}

// Basic processes basic authentication logic.
func Basic(userBus *userbus.Business, ath *auth.Auth) web.MidHandler {
	midFunc := func(ctx context.Context, w http.ResponseWriter, r *http.Request, hdl mid.Handler) (any, error) {
		return mid.Basic(ctx, ath, userBus, r.Header.Get("authorization"), hdl)
	}

	return addMiddleware(midFunc)
}
