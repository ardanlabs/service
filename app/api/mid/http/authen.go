package http

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/app/api/mid"
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/core/crud/userbus"
	"github.com/ardanlabs/service/foundation/authapi"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// AuthenticateWeb validates a JWT from the `Authorization` header.
func AuthenticateWeb(log *logger.Logger, authAPI *authapi.AuthAPI) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()

			resp, err := authAPI.Authenticate(ctx, r.Header.Get("authorization"))
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			ctx = mid.SetUserID(ctx, resp.UserID)
			ctx = mid.SetClaims(ctx, resp.Claims)

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

// Authenticate validates a JWT from the `Authorization` header.
func Authenticate(userBus *userbus.Core, auth *auth.Auth) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			authorization := r.Header.Get("authorization")
			parts := strings.Split(authorization, " ")
			if len(parts) != 2 {
				return errs.Newf(errs.Unauthenticated, "invalid authorization value")
			}

			var err error

			switch parts[0] {
			case "Bearer":
				ctx, err = processJWT(ctx, auth, authorization)

			case "Basic":
				ctx, err = processBasic(ctx, userBus, authorization)
			}

			if err != nil {
				return err
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

func processJWT(ctx context.Context, auth *auth.Auth, token string) (context.Context, error) {
	claims, err := auth.Authenticate(ctx, token)
	if err != nil {
		return ctx, errs.New(errs.Unauthenticated, err)
	}

	if claims.Subject == "" {
		return ctx, errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, no claims")
	}

	subjectID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return ctx, errs.New(errs.Unauthenticated, fmt.Errorf("parsing subject: %w", err))
	}

	ctx = mid.SetUserID(ctx, subjectID)
	ctx = mid.SetClaims(ctx, claims)

	return ctx, nil
}

func processBasic(ctx context.Context, userBus *userbus.Core, basic string) (context.Context, error) {
	email, pass, ok := parseBasicAuth(basic)
	if !ok {
		return ctx, errs.Newf(errs.Unauthenticated, "invalid Basic auth")
	}

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return ctx, errs.New(errs.Unauthenticated, err)
	}

	usr, err := userBus.Authenticate(ctx, *addr, pass)
	if err != nil {
		return ctx, errs.New(errs.Unauthenticated, err)
	}

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   usr.ID.String(),
			Issuer:    "service project",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(8760 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: usr.Roles,
	}

	subjectID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return ctx, errs.Newf(errs.Unauthenticated, "parsing subject: %s", err)
	}

	ctx = mid.SetUserID(ctx, subjectID)
	ctx = mid.SetClaims(ctx, claims)

	return ctx, nil
}

func parseBasicAuth(auth string) (string, string, bool) {
	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Basic" {
		return "", "", false
	}

	c, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", false
	}

	username, password, ok := strings.Cut(string(c), ":")
	if !ok {
		return "", "", false
	}

	return username, password, true
}
