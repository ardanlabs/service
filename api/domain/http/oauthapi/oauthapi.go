// Package oauthapi maintains the web based api for oauth support.
package oauthapi

import (
	"net/http"
	"time"

	"github.com/ardanlabs/service/app/sdk/auth"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

type api struct {
	log             *logger.Logger
	auth            *auth.Auth
	store           sessions.Store
	tokenKey        string
	uiAdminRedirect string
	uiLoginRedirect string
}

func newAPI(cfg Config) *api {
	goth.UseProviders(
		google.New(cfg.GoogleKey, cfg.GoogleSecret, cfg.Callback),
	)

	gothic.GetProviderName = func(r *http.Request) (string, error) {
		return "google", nil
	}

	store := sessions.NewCookieStore([]byte(cfg.StoreKey))
	store.Options = &sessions.Options{
		Path: "/",
	}

	return &api{
		auth:            cfg.Auth,
		store:           store,
		tokenKey:        cfg.TokenKey,
		uiAdminRedirect: cfg.UIAdminRedirect,
		uiLoginRedirect: cfg.UILoginRedirect,
	}
}

func (a *api) authenticate(w http.ResponseWriter, r *http.Request) {
	gothic.BeginAuthHandler(w, r)
}

func (a *api) authCallback(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		a.log.Error(r.Context(), "completing user auth: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sess, err := a.store.Get(r, "user-metadata")
	if err != nil {
		a.log.Error(r.Context(), "get session: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sess.Values["user"] = user

	if err := sess.Save(r, w); err != nil {
		a.log.Error(r.Context(), "save session: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.UserID,
			Issuer:    a.auth.Issuer(),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(20 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: []string{userbus.Roles.Admin.String()},
	}

	token, err := a.auth.GenerateToken(a.tokenKey, claims)
	if err != nil {
		a.log.Error(r.Context(), "generating token: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, a.uiAdminRedirect+token, http.StatusFound)
}

func (a *api) logout(w http.ResponseWriter, r *http.Request) {
	sess, err := a.store.Get(r, "user-metadata")
	if err != nil {
		a.log.Error(r.Context(), "get session: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Logout user by invalidating their session data.
	sess.Values["user"] = nil

	if err := sess.Save(r, w); err != nil {
		a.log.Error(r.Context(), "save session: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := gothic.Logout(w, r); err != nil {
		a.log.Error(r.Context(), "gothic logout: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, a.uiLoginRedirect, http.StatusFound)
}
