package mid

import (
	"context"
	"crypto/rsa"
	"log"
	"net/http"
	"strings"

	"github.com/ardanlabs/service/internal/auth"
	"github.com/ardanlabs/service/internal/platform/web"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

// KeyFunc is used to map a JWT key id (kid) to the corresponding public key.
type KeyFunc func(kid string) (*rsa.PublicKey, error)

// Auth is used to authenticate HTTP requests.
type Auth struct {
	KeyFunc KeyFunc
}

// Authenticate validates a JWT from the `Authorization` header.
func (a *Auth) Authenticate(next web.Handler) web.Handler {
	h := func(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
		tknStr, err := getBearerToken(r)
		if err != nil {
			return errors.Wrap(web.ErrUnauthorized, err.Error())
		}

		var claims auth.Claims
		tkn, err := jwt.ParseWithClaims(tknStr, &claims, func(t *jwt.Token) (interface{}, error) {
			kid, ok := t.Header["kid"]
			if !ok {
				return nil, errors.New("Missing key id (kid) in token header")
			}
			kidStr, ok := kid.(string)
			if !ok {
				return nil, errors.New("Token key id (kid) must be string")
			}

			return a.KeyFunc(kidStr)
		})

		if !tkn.Valid {
			return errors.Wrap(web.ErrUnauthorized, "Invalid token")
		}

		// TODO: Do we need to assert claims?

		// TODO: Pass claims down.

		return next(ctx, log, w, r, params)
	}

	return h
}

// getBearerToken grabs a token from the request header. Expected header is of
// the format `Authorization: Bearer <token>`.
func getBearerToken(r *http.Request) (string, error) {
	hdr := r.Header.Get("Authorization")
	if hdr == "" {
		return "", errors.New("Missing Authorization header")
	}

	split := strings.Split(hdr, " ")
	if len(split) != 2 || strings.ToLower(split[0]) != "bearer" {
		return "", errors.New("Expected Authorization header format: Bearer <token>")
	}

	return split[1], nil
}
