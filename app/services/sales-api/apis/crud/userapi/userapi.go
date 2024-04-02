// Package userapi maintains the group of handlers for user access.
package userapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"time"

	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/api/errs"
	"github.com/ardanlabs/service/business/api/mid"
	"github.com/ardanlabs/service/business/api/page"
	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/golang-jwt/jwt/v4"
)

// API manages the set of handler functions for this domain.
type API struct {
	user *user.Core
	auth *auth.Auth
}

// New constructs a Handlers for use.
func New(user *user.Core, auth *auth.Auth) *API {
	return &API{
		user: user,
		auth: auth,
	}
}

// Token provides an API token for the authenticated user.
func (api *API) Token(ctx context.Context, kid string, addr mail.Address, password string) (Token, error) {
	usr, err := api.user.Authenticate(ctx, addr, password)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrNotFound):
			return Token{}, errs.New(http.StatusBadRequest, err)
		case errors.Is(err, user.ErrAuthenticationFailure):
			return Token{}, errs.New(http.StatusUnauthorized, err)
		default:
			return Token{}, fmt.Errorf("authenticate: %w", err)
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

	tkn, err := api.auth.GenerateToken(kid, claims)
	if err != nil {
		return Token{}, errs.New(http.StatusInternalServerError, err)
	}

	return toToken(tkn), nil
}

// Create adds a new user to the system.
func (api *API) Create(ctx context.Context, app AppNewUser) (AppUser, error) {
	nc, err := toCoreNewUser(app)
	if err != nil {
		return AppUser{}, errs.New(http.StatusBadRequest, err)
	}

	usr, err := api.user.Create(ctx, nc)
	if err != nil {
		if errors.Is(err, user.ErrUniqueEmail) {
			return AppUser{}, errs.New(http.StatusConflict, user.ErrUniqueEmail)
		}
		return AppUser{}, errs.Newf(http.StatusInternalServerError, "create: usr[%+v]: %s", usr, err)
	}

	return toAppUser(usr), nil
}

// Update updates an existing user.
func (api *API) Update(ctx context.Context, app AppUpdateUser) (AppUser, error) {
	uu, err := toCoreUpdateUser(app)
	if err != nil {
		return AppUser{}, errs.New(http.StatusBadRequest, err)
	}

	usr, err := mid.GetUser(ctx)
	if err != nil {
		return AppUser{}, errs.Newf(http.StatusInternalServerError, "user missing in context: %s", err)
	}

	updUsr, err := api.user.Update(ctx, usr, uu)
	if err != nil {
		return AppUser{}, errs.Newf(http.StatusInternalServerError, "update: userID[%s] uu[%+v]: %s", usr.ID, uu, err)
	}

	return toAppUser(updUsr), nil
}

// UpdateRole updates an existing user's role.
func (api *API) UpdateRole(ctx context.Context, app AppUpdateUserRole) (AppUser, error) {
	uu, err := toCoreUpdateUserRole(app)
	if err != nil {
		return AppUser{}, errs.New(http.StatusBadRequest, err)
	}

	usr, err := mid.GetUser(ctx)
	if err != nil {
		return AppUser{}, errs.Newf(http.StatusInternalServerError, "user missing in context: %s", err)
	}

	updUsr, err := api.user.Update(ctx, usr, uu)
	if err != nil {
		return AppUser{}, errs.Newf(http.StatusInternalServerError, "updaterole: userID[%s] uu[%+v]: %s", usr.ID, uu, err)
	}

	return toAppUser(updUsr), nil
}

// Delete removes a user from the system.
func (api *API) Delete(ctx context.Context) error {
	usr, err := mid.GetUser(ctx)
	if err != nil {
		return errs.Newf(http.StatusInternalServerError, "userID missing in context: %s", err)
	}

	if err := api.user.Delete(ctx, usr); err != nil {
		return errs.Newf(http.StatusInternalServerError, "delete: userID[%s]: %s", usr.ID, err)
	}

	return nil
}

// Query returns a list of users with paging.
func (api *API) Query(ctx context.Context, qp QueryParams) (page.Document[AppUser], error) {
	if err := validatePaging(qp); err != nil {
		return page.Document[AppUser]{}, err
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return page.Document[AppUser]{}, err
	}

	orderBy, err := parseOrder(qp)
	if err != nil {
		return page.Document[AppUser]{}, err
	}

	usrs, err := api.user.Query(ctx, filter, orderBy, qp.Page, qp.Rows)
	if err != nil {
		return page.Document[AppUser]{}, errs.Newf(http.StatusInternalServerError, "query: %s", err)
	}

	total, err := api.user.Count(ctx, filter)
	if err != nil {
		return page.Document[AppUser]{}, errs.Newf(http.StatusInternalServerError, "count: %s", err)
	}

	return page.NewDocument(toAppUsers(usrs), total, qp.Page, qp.Rows), nil
}

// QueryByID returns a user by its ID.
func (api *API) QueryByID(ctx context.Context) (AppUser, error) {
	usr, err := mid.GetUser(ctx)
	if err != nil {
		return AppUser{}, errs.Newf(http.StatusInternalServerError, "querybyid: %s", err)
	}

	return toAppUser(usr), nil
}
