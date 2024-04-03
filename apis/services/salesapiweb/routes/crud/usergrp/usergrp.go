// Package usergrp maintains the group of handlers for user access.
package usergrp

import (
	"context"
	"errors"
	"net/http"
	"net/mail"

	"github.com/ardanlabs/service/app/core/crud/userapp"
	"github.com/ardanlabs/service/business/api/errs"
	"github.com/ardanlabs/service/foundation/validate"
	"github.com/ardanlabs/service/foundation/web"
)

type handlers struct {
	user *userapp.Core
}

func new(user *userapp.Core) *handlers {
	return &handlers{
		user: user,
	}
}

func (h *handlers) create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app userapp.NewUser
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	usr, err := h.user.Create(ctx, app)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, usr, http.StatusCreated)
}

func (h *handlers) update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app userapp.UpdateUser
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	usr, err := h.user.Update(ctx, app)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, usr, http.StatusOK)
}

func (h *handlers) updateRole(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app userapp.UpdateUserRole
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	usr, err := h.user.UpdateRole(ctx, app)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, usr, http.StatusOK)
}

func (h *handlers) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if err := h.user.Delete(ctx); err != nil {
		return err
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *handlers) query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	qp, err := parseQueryParams(r)
	if err != nil {
		return err
	}

	usr, err := h.user.Query(ctx, qp)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, usr, http.StatusOK)
}

func (h *handlers) queryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	usr, err := h.user.QueryByID(ctx)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, usr, http.StatusOK)
}

func (h *handlers) token(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	kid := web.Param(r, "kid")
	if kid == "" {
		return validate.NewFieldsError("kid", errors.New("missing kid"))
	}

	email, pass, ok := r.BasicAuth()
	if !ok {
		return errs.Newf(errs.Unauthenticated, "authorize: must provide email and password in Basic auth")
	}

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return errs.Newf(errs.Unauthenticated, "authorize: invalid email format")
	}

	usr, err := h.user.Token(ctx, kid, *addr, pass)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, usr, http.StatusOK)
}
