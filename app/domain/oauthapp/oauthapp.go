// Package oauthapp maintains the web based api for oauth support.
package oauthapp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ardanlabs/service/app/sdk/auth"
	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/business/types/role"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/golang-jwt/jwt/v4"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

type app struct {
	log      *logger.Logger
	auth     *auth.Auth
	tokenKey string
	uiURL    string
	apiHost  string
}

func newApp(cfg Config) *app {
	goth.UseProviders(
		google.New(cfg.GoogleKey, cfg.GoogleSecret, fmt.Sprintf("%s/api/auth/google/callback", cfg.GoogleCallBackURL)),
	)

	gothic.GetProviderName = func(r *http.Request) (string, error) {
		return web.Param(r, "provider"), nil
	}

	return &app{
		auth:    cfg.Auth,
		log:     cfg.Log,
		uiURL:   cfg.GoogleUIURL,
		apiHost: cfg.APIHost,
	}
}

func (a *app) authenticate(ctx context.Context, r *http.Request) web.Encoder {
	gothic.BeginAuthHandler(web.GetWriter(ctx), r)

	return web.NewNoResponse()
}

func (a *app) authCallback(ctx context.Context, r *http.Request) web.Encoder {
	w := web.GetWriter(ctx)

	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	clms := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.UserID,
			Issuer:    a.auth.Issuer(),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(2 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: []string{role.Admin.String()},
	}

	token, err := a.auth.GenerateToken(a.tokenKey, clms)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	redirect := fmt.Sprintf("%s/app/admin?token=%s", a.uiURL, token)
	a.log.Info(r.Context(), "REDIRECT", "redirect", redirect)

	http.Redirect(w, r, redirect, http.StatusFound)

	return web.NewNoResponse()
}

func (a *app) logout(ctx context.Context, r *http.Request) web.Encoder {
	w := web.GetWriter(ctx)

	if err := gothic.Logout(w, r); err != nil {
		return errs.New(errs.Internal, err)
	}

	redirect := "/app/login"
	http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)

	return web.NewNoResponse()
}
