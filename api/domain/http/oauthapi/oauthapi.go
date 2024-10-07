// Package oauthapi maintains the web based api for oauth support.
package oauthapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ardanlabs/service/app/sdk/auth"
	"github.com/ardanlabs/service/business/types/role"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/golang-jwt/jwt/v4"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

type api struct {
	log      *logger.Logger
	auth     *auth.Auth
	tokenKey string
	uiURL    string
	apiHost  string
}

func newAPI(cfg Config) *api {
	goth.UseProviders(
		google.New(cfg.GoogleKey, cfg.GoogleSecret, fmt.Sprintf("%s/api/auth/google/callback", cfg.GoogleCallBackURL)),
	)

	gothic.GetProviderName = func(r *http.Request) (string, error) {
		return web.Param(r, "provider"), nil
	}

	return &api{
		auth:    cfg.Auth,
		log:     cfg.Log,
		uiURL:   cfg.GoogleUIURL,
		apiHost: cfg.APIHost,
	}
}

func (a *api) authenticate(w http.ResponseWriter, r *http.Request) {
	gothic.BeginAuthHandler(w, r)
}

func (a *api) authCallback(w http.ResponseWriter, r *http.Request) {
	var user goth.User

	var err error
	user, err = gothic.CompleteUserAuth(w, r)
	if err != nil {
		a.log.Error(r.Context(), "completing user auth", "msg", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
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
		a.log.Error(r.Context(), "generating token", "msg", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	redirect := fmt.Sprintf("%s/app/admin?token=%s", a.uiURL, token)
	a.log.Info(r.Context(), "REDIRECT", "redirect", redirect)

	http.Redirect(w, r, redirect, http.StatusFound)
}

func (a *api) logout(w http.ResponseWriter, r *http.Request) {
	if err := gothic.Logout(w, r); err != nil {
		a.log.Error(r.Context(), "gothic logout", "msg", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	redirect := "/app/login"
	http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
}
