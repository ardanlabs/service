package mid

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/ardanlabs/service/internal/auth"
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
		tknStr, err := getBearerToken(r.Header.Get("Authorization"))
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

// getBearerToken grabs a token from the request header. Expected header is of
// the format `Authorization: Bearer <token>`.
func getBearerToken(hdr string) (string, error) {
	if hdr == "" {
		return "", errors.New("Missing Authorization header")
	}

	split := strings.Split(hdr, " ")
	if len(split) != 2 || strings.ToLower(split[0]) != "bearer" {
		return "", errors.New("Expected Authorization header format: Bearer <token>")
	}

	return split[1], nil
}
