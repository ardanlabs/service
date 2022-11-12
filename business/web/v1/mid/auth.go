package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/business/web/auth"
	"github.com/ardanlabs/service/foundation/web"
)

// Authenticate validates a JWT from the `Authorization` header.
func Authenticate(a *auth.Auth) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			claims, err := a.Authenticate(ctx, r.Header.Get("authorization"))
			if err != nil {
				return auth.NewAuthErrorf("authenticate: failed: %s", err)
			}

			ctx = auth.SetClaims(ctx, claims)

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

// Authorize validates that an authenticated user has at least one role from a
// specified list. This method constructs the actual function that is used.
func Authorize(a *auth.Auth, roles ...string) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			claims := auth.GetClaims(ctx)
			if claims.Subject == "" {
				return auth.NewAuthErrorf("authorize: you are not authorized for that action, no claims")
			}

			if err := a.Authorize(ctx, claims, roles...); err != nil {
				return auth.NewAuthErrorf("authorize: you are not authorized for that action, claims[%v] roles[%v]: %s", claims.Roles, roles, err)
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
