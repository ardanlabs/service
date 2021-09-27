// Package usergrp maintains the group of handlers for user access.
package usergrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/sys/auth"
	"github.com/ardanlabs/service/business/sys/validate"
	webv1 "github.com/ardanlabs/service/business/web/v1"
	"github.com/ardanlabs/service/foundation/web"
)

// Handlers manages the set of user enpoints.
type Handlers struct {
	Core user.Core
	Auth *auth.Auth
}

// Create adds a new user to the system.
func (h Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	var nu user.NewUser
	if err := web.Decode(r, &nu); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	usr, err := h.Core.Create(ctx, nu, v.Now)
	if err != nil {
		return fmt.Errorf("user[%+v]: %w", &usr, err)
	}

	return web.Respond(ctx, w, usr, http.StatusCreated)
}

// Update updates a user in the system.
func (h Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return webv1.NewRequestError(auth.ErrForbidden, http.StatusForbidden)
	}

	var upd user.UpdateUser
	if err := web.Decode(r, &upd); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	userID := web.Param(r, "id")

	// If you are not an admin and looking to retrieve someone other than yourself.
	if !claims.Authorized(auth.RoleAdmin) && claims.Subject != userID {
		return webv1.NewRequestError(auth.ErrForbidden, http.StatusForbidden)
	}

	if err := h.Core.Update(ctx, userID, upd, v.Now); err != nil {
		switch validate.Cause(err) {
		case validate.ErrInvalidID:
			return webv1.NewRequestError(err, http.StatusBadRequest)
		case validate.ErrNotFound:
			return webv1.NewRequestError(err, http.StatusNotFound)
		case auth.ErrForbidden:
			return webv1.NewRequestError(err, http.StatusForbidden)
		default:
			return fmt.Errorf("ID[%s] User[%+v]: %w", userID, &upd, err)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Delete removes a user from the system.
func (h Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return webv1.NewRequestError(auth.ErrForbidden, http.StatusForbidden)
	}

	userID := web.Param(r, "id")

	// If you are not an admin and looking to delete someone other than yourself.
	if !claims.Authorized(auth.RoleAdmin) && claims.Subject != userID {
		return webv1.NewRequestError(auth.ErrForbidden, http.StatusForbidden)
	}

	if err := h.Core.Delete(ctx, userID); err != nil {
		switch validate.Cause(err) {
		case validate.ErrInvalidID:
			return webv1.NewRequestError(err, http.StatusBadRequest)
		case validate.ErrNotFound:
			return webv1.NewRequestError(err, http.StatusNotFound)
		case auth.ErrForbidden:
			return webv1.NewRequestError(err, http.StatusForbidden)
		default:
			return fmt.Errorf("ID[%s]: %w", userID, err)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Query returns a list of users with paging.
func (h Handlers) Query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page := web.Param(r, "page")
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		return webv1.NewRequestError(fmt.Errorf("invalid page format [%s]", page), http.StatusBadRequest)
	}
	rows := web.Param(r, "rows")
	rowsPerPage, err := strconv.Atoi(rows)
	if err != nil {
		return webv1.NewRequestError(fmt.Errorf("invalid rows format [%s]", rows), http.StatusBadRequest)
	}

	users, err := h.Core.Query(ctx, pageNumber, rowsPerPage)
	if err != nil {
		return fmt.Errorf("unable to query for users: %w", err)
	}

	return web.Respond(ctx, w, users, http.StatusOK)
}

// QueryByID returns a user by its ID.
func (h Handlers) QueryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return webv1.NewRequestError(auth.ErrForbidden, http.StatusForbidden)
	}

	userID := web.Param(r, "id")

	// If you are not an admin and looking to retrieve someone other than yourself.
	if !claims.Authorized(auth.RoleAdmin) && claims.Subject != userID {
		return webv1.NewRequestError(auth.ErrForbidden, http.StatusForbidden)
	}

	usr, err := h.Core.QueryByID(ctx, userID)
	if err != nil {
		switch validate.Cause(err) {
		case validate.ErrInvalidID:
			return webv1.NewRequestError(err, http.StatusBadRequest)
		case validate.ErrNotFound:
			return webv1.NewRequestError(err, http.StatusNotFound)
		case auth.ErrForbidden:
			return webv1.NewRequestError(err, http.StatusForbidden)
		default:
			return fmt.Errorf("ID[%s]: %w", userID, err)
		}
	}

	return web.Respond(ctx, w, usr, http.StatusOK)
}

// Token provides an API token for the authenticated user.
func (h Handlers) Token(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	email, pass, ok := r.BasicAuth()
	if !ok {
		err := errors.New("must provide email and password in Basic auth")
		return webv1.NewRequestError(err, http.StatusUnauthorized)
	}

	claims, err := h.Core.Authenticate(ctx, v.Now, email, pass)
	if err != nil {
		switch validate.Cause(err) {
		case validate.ErrNotFound:
			return webv1.NewRequestError(err, http.StatusNotFound)
		case auth.ErrAuthenticationFailure:
			return webv1.NewRequestError(err, http.StatusUnauthorized)
		default:
			return fmt.Errorf("authenticating: %w", err)
		}
	}

	var tkn struct {
		Token string `json:"token"`
	}
	tkn.Token, err = h.Auth.GenerateToken(claims)
	if err != nil {
		return fmt.Errorf("generating token: %w", err)
	}

	return web.Respond(ctx, w, tkn, http.StatusOK)
}
