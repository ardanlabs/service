package mid

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/ardanlabs/service/internal/platform/auth"
	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

// Authenticate validates a JWT from the `Authorization` header.
func (mw *Middleware) Authenticate(after web.Handler) web.Handler {

	// Wrap this handler around the next one provided.
	h := func(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
		ctx, span := trace.StartSpan(ctx, "internal.mid.Authenticate")
		defer span.End()

		authHdr := r.Header.Get("Authorization")
		if authHdr == "" {
			err := errors.New("missing Authorization header")
			return web.WrapErrorWithStatus(err, http.StatusUnauthorized)
		}

		tknStr, err := parseAuthHeader(authHdr)
		if err != nil {
			return web.WrapErrorWithStatus(err, http.StatusUnauthorized)
		}

		claims, err := mw.Authenticator.ParseClaims(tknStr)
		if err != nil {
			return web.WrapErrorWithStatus(err, http.StatusUnauthorized)
		}

		// Add claims to the context so they can be retrieved later.
		ctx = context.WithValue(ctx, auth.Key, claims)

		return after(ctx, log, w, r, params)
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

// ErrForbidden is returned when an authenticated user does not have a
// sufficient role for an action.
var ErrForbidden = web.WrapErrorWithStatus(
	errors.New("you are not authorized for that action"),
	http.StatusUnauthorized,
)

// HasRole validates that an authenticated user has at least one role from a
// specified list. This method constructs the actual function that is used.
func (mw *Middleware) HasRole(roles ...string) func(next web.Handler) web.Handler {
	fn := func(next web.Handler) web.Handler {
		h := func(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
			ctx, span := trace.StartSpan(ctx, "internal.mid.HasRole")
			defer span.End()

			claims, ok := ctx.Value(auth.Key).(auth.Claims)
			if !ok {
				// TODO(jlw) should this be a web.Shutdown?
				return errors.New("claims missing from context: HasRole called without/before Authenticate")
			}

			if !claims.HasRole(roles...) {
				return ErrForbidden
			}

			return next(ctx, log, w, r, params)
		}

		return h
	}

	return fn
}
