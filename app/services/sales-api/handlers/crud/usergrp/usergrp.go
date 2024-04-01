// Package usergrp maintains the group of handlers for user access.
package usergrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"time"

	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/web/auth"
	"github.com/ardanlabs/service/business/web/errs"
	"github.com/ardanlabs/service/business/web/mid"
	"github.com/ardanlabs/service/business/web/page"
	"github.com/ardanlabs/service/foundation/validate"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/golang-jwt/jwt/v4"
)

type handlers struct {
	user *user.Core
	auth *auth.Auth
}

func new(user *user.Core, auth *auth.Auth) *handlers {
	return &handlers{
		user: user,
		auth: auth,
	}
}

// create adds a new user to the system.
func (h *handlers) create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app AppNewUser
	if err := web.Decode(r, &app); err != nil {
		return errs.NewTrusted(err, http.StatusBadRequest)
	}

	nc, err := toCoreNewUser(app)
	if err != nil {
		return errs.NewTrusted(err, http.StatusBadRequest)
	}

	usr, err := h.user.Create(ctx, nc)
	if err != nil {
		if errors.Is(err, user.ErrUniqueEmail) {
			return errs.NewTrusted(err, http.StatusConflict)
		}
		return fmt.Errorf("create: usr[%+v]: %w", usr, err)
	}

	return web.Respond(ctx, w, toAppUser(usr), http.StatusCreated)
}

// update updates a user in the system.
func (h *handlers) update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app AppUpdateUser
	if err := web.Decode(r, &app); err != nil {
		return errs.NewTrusted(err, http.StatusBadRequest)
	}

	uu, err := toCoreUpdateUser(app)
	if err != nil {
		return errs.NewTrusted(err, http.StatusBadRequest)
	}

	usr, err := mid.GetUser(ctx)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}

	updUsr, err := h.user.Update(ctx, usr, uu)
	if err != nil {
		return fmt.Errorf("update: userID[%s] uu[%+v]: %w", usr.ID, uu, err)
	}

	return web.Respond(ctx, w, toAppUser(updUsr), http.StatusOK)
}

// updateRole updates a user role in the system.
func (h *handlers) updateRole(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app AppUpdateUserRole
	if err := web.Decode(r, &app); err != nil {
		return errs.NewTrusted(err, http.StatusBadRequest)
	}

	uu, err := toCoreUpdateUserRole(app)
	if err != nil {
		return errs.NewTrusted(err, http.StatusBadRequest)
	}

	usr, err := mid.GetUser(ctx)
	if err != nil {
		return fmt.Errorf("updaterole: %w", err)
	}

	updUsr, err := h.user.Update(ctx, usr, uu)
	if err != nil {
		return fmt.Errorf("updaterole: userID[%s] uu[%+v]: %w", usr.ID, uu, err)
	}

	return web.Respond(ctx, w, toAppUser(updUsr), http.StatusOK)
}

// delete removes a user from the system.
func (h *handlers) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	usr, err := mid.GetUser(ctx)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	if err := h.user.Delete(ctx, usr); err != nil {
		return fmt.Errorf("delete: userID[%s]: %w", usr.ID, err)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// query returns a list of users with paging.
func (h *handlers) query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	pg, err := page.Parse(r)
	if err != nil {
		return err
	}

	filter, err := parseFilter(r)
	if err != nil {
		return err
	}

	orderBy, err := parseOrder(r)
	if err != nil {
		return err
	}

	users, err := h.user.Query(ctx, filter, orderBy, pg.Number, pg.RowsPerPage)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	total, err := h.user.Count(ctx, filter)
	if err != nil {
		return fmt.Errorf("count: %w", err)
	}

	return web.Respond(ctx, w, page.NewDocument(toAppUsers(users), total, pg.Number, pg.RowsPerPage), http.StatusOK)
}

// queryByID returns a user by its ID.
func (h *handlers) queryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	usr, err := mid.GetUser(ctx)
	if err != nil {
		return fmt.Errorf("querybyid: %w", err)
	}

	return web.Respond(ctx, w, toAppUser(usr), http.StatusOK)
}

// token provides an API token for the authenticated user.
func (h *handlers) token(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	kid := web.Param(r, "kid")
	if kid == "" {
		return validate.NewFieldsError("kid", errors.New("missing kid"))
	}

	email, pass, ok := r.BasicAuth()
	if !ok {
		return auth.NewAuthError("must provide email and password in Basic auth")
	}

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return auth.NewAuthError("invalid email format")
	}

	usr, err := h.user.Authenticate(ctx, *addr, pass)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrNotFound):
			return errs.NewTrusted(err, http.StatusNotFound)
		case errors.Is(err, user.ErrAuthenticationFailure):
			return auth.NewAuthError(err.Error())
		default:
			return fmt.Errorf("authenticate: %w", err)
		}
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

	token, err := h.auth.GenerateToken(kid, claims)
	if err != nil {
		return fmt.Errorf("generatetoken: %w", err)
	}

	return web.Respond(ctx, w, toToken(token), http.StatusOK)
}
