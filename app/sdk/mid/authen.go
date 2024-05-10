package mid

import (
	"context"
	"encoding/base64"
	"net/mail"
	"strings"
	"time"

	"github.com/ardanlabs/service/app/sdk/auth"
	"github.com/ardanlabs/service/app/sdk/authclient"
	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// Authenticate validates authentication via the auth service.
func Authenticate(ctx context.Context, log *logger.Logger, client *authclient.Client, authorization string, next Handler) (Encoder, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := client.Authenticate(ctx, authorization)
	if err != nil {
		return nil, errs.New(errs.Unauthenticated, err)
	}

	ctx = setUserID(ctx, resp.UserID)
	ctx = setClaims(ctx, resp.Claims)

	return next(ctx)
}

// Bearer processes JWT authentication logic.
func Bearer(ctx context.Context, ath *auth.Auth, authorization string, next Handler) (Encoder, error) {
	claims, err := ath.Authenticate(ctx, authorization)
	if err != nil {
		return nil, errs.New(errs.Unauthenticated, err)
	}

	if claims.Subject == "" {
		return nil, errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, no claims")
	}

	subjectID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, errs.Newf(errs.Unauthenticated, "parsing subject: %s", err)
	}

	ctx = setUserID(ctx, subjectID)
	ctx = setClaims(ctx, claims)

	return next(ctx)
}

// Basic processes basic authentication logic.
func Basic(ctx context.Context, ath *auth.Auth, userBus *userbus.Business, authorization string, next Handler) (Encoder, error) {
	email, pass, ok := parseBasicAuth(authorization)
	if !ok {
		return nil, errs.Newf(errs.Unauthenticated, "invalid Basic auth")
	}

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return nil, errs.New(errs.Unauthenticated, err)
	}

	usr, err := userBus.Authenticate(ctx, *addr, pass)
	if err != nil {
		return nil, errs.New(errs.Unauthenticated, err)
	}

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   usr.ID.String(),
			Issuer:    ath.Issuer(),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(8760 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: usr.Roles,
	}

	subjectID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, errs.Newf(errs.Unauthenticated, "parsing subject: %s", err)
	}

	ctx = setUserID(ctx, subjectID)
	ctx = setClaims(ctx, claims)

	return next(ctx)
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
