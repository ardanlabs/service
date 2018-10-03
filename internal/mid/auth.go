package mid

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/ardanlabs/service/internal/platform/auth"
	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/pkg/errors"
)

// Auth is used to authenticate and authorize HTTP requests.
type Auth struct {
	Authenticator *auth.Authenticator
}

// Authenticate validates a JWT from the `Authorization` header.
func (a *Auth) Authenticate(next web.Handler) web.Handler {
	h := func(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
		authHdr := r.Header.Get("Authorization")
		if authHdr == "" {
			return errors.Wrap(web.ErrUnauthorized, "Missing Authorization header")
		}

		tknStr, err := parseAuthHeader(authHdr)
		if err != nil {
			return errors.Wrap(web.ErrUnauthorized, err.Error())
		}

		claims, err := a.Authenticator.ParseClaims(tknStr)
		if err != nil {
			return errors.Wrap(web.ErrUnauthorized, err.Error())
		}

		// Add claims to the context so they can be retrieved later.
		ctx = context.WithValue(ctx, auth.Key, claims)

		return next(ctx, log, w, r, params)
	}

	return h
}

// parseAuthHeader parses an authorization header. Expected header is of
// the format `Bearer <token>`.
func parseAuthHeader(bearerStr string) (string, error) {
	split := strings.Split(bearerStr, " ")
	if len(split) != 2 || strings.ToLower(split[0]) != "bearer" {
		return "", errors.New("Expected Authorization header format: Bearer <token>")
	}

	return split[1], nil
}

// HasRole validates that an authenticated user has at least one role from a
// specified list. This method constructs the actual function that is used.
func (a *Auth) HasRole(roles ...string) func(next web.Handler) web.Handler {
	mw := func(next web.Handler) web.Handler {
		h := func(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {

			claims, ok := ctx.Value(auth.Key).(auth.Claims)
			if !ok {
				return web.ErrUnauthorized
			}

			if !claims.HasRole(roles...) {
				return web.ErrForbidden
			}

			return next(ctx, log, w, r, params)
		}

		return h
	}
	return mw
}
