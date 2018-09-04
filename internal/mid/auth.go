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

// Auth is used to authenticate HTTP requests.
type Auth struct {
	Parser *auth.Parser
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

		claims, err := a.Parser.ParseClaims(tknStr)
		if err != nil {
			return errors.Wrap(web.ErrUnauthorized, err.Error())
		}

		// TODO: Pass claims down.
		// Options:
		// A. Pass via context
		// B. Pass via additional param
		_ = claims

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
