package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/apis/api/authclient"
	"github.com/ardanlabs/service/app/api/mid"
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// Authenticate validates a JWT from the `Authorization` header.
func Authenticate(log *logger.Logger, client *authclient.Client) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx, err := mid.Authenticate(ctx, log, client, r.Header.Get("authorization"))
			if err != nil {
				return err
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

// Authorization validates a JWT from the `Authorization` header.
func Authorization(userBus *userbus.Core, auth *auth.Auth) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			authorization := r.Header.Get("authorization")

			ctx, err := mid.Authorization(ctx, userBus, auth, authorization)
			if err != nil {
				return err
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
