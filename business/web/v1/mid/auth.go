package mid

import (
	"context"
	"errors"
	"net/http"

	"github.com/ardanlabs/service/business/web/auth"
	v1 "github.com/ardanlabs/service/business/web/v1"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/google/uuid"
)

// Set of error variables for handling user group errors.
var (
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// Authenticate validates a JWT from the `Authorization` header.
func Authenticate(a *auth.Auth) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			claims, err := a.Authenticate(ctx, r.Header.Get("authorization"))
			if err != nil {
				return auth.NewAuthError("authenticate: failed: %s", err)
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
func Authorize(a *auth.Auth, rule string) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			claims := auth.GetClaims(ctx)
			if claims.Subject == "" {
				return auth.NewAuthError("authorize: you are not authorized for that action, no claims")
			}

			// I will use an zero valued user id if it doesn't exsit.
			var userID uuid.UUID
			id := web.Param(r, "user_id")
			if id != "" {
				var err error
				userID, err = uuid.Parse(id)
				if err != nil {
					return v1.NewRequestError(ErrInvalidID, http.StatusBadRequest)
				}
				ctx = auth.SetUserID(ctx, userID)
			}

			if err := a.Authorize(ctx, claims, userID, rule); err != nil {
				return auth.NewAuthError("authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", claims.Roles, rule, err)
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
