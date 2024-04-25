package mid

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/ardanlabs/service/app/api/authclient"
	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// Authenticate validates a JWT from the `Authorization` header.
func Authenticate(ctx context.Context, log *logger.Logger, client *authclient.Client, authorization string, handler Handler) error {
	resp, err := client.Authenticate(ctx, authorization)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	ctx = setUserID(ctx, resp.UserID)
	ctx = setClaims(ctx, resp.Claims)

	return handler(ctx)
}

// Authorization validates a JWT from the `Authorization` header.
func Authorization(ctx context.Context, userBus *userbus.Core, auth *auth.Auth, authorization string, handler Handler) error {
	var err error
	parts := strings.Split(authorization, " ")

	switch parts[0] {
	case "Bearer":
		ctx, err = processJWT(ctx, auth, authorization)

	case "Basic":
		ctx, err = processBasic(ctx, userBus, authorization)
	}

	if err != nil {
		return err
	}

	return handler(ctx)
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

	ctx = setUserID(ctx, subjectID)
	ctx = setClaims(ctx, claims)

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

	ctx = setUserID(ctx, subjectID)
	ctx = setClaims(ctx, claims)

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
