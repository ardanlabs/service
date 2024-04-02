// Package userapp maintains the app layer api for the user domain.
package userapp

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/api/errs"
	"github.com/ardanlabs/service/business/api/mid"
	"github.com/ardanlabs/service/business/api/page"
	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/golang-jwt/jwt/v4"
)

// Core manages the set of handler functions for this domain.
type Core struct {
	user *user.Core
	auth *auth.Auth
}

// New constructs a Handlers for use.
func New(user *user.Core, auth *auth.Auth) *Core {
	return &Core{
		user: user,
		auth: auth,
	}
}

// Token provides an API token for the authenticated user.
func (c *Core) Token(ctx context.Context, kid string, addr mail.Address, password string) (Token, error) {
	usr, err := c.user.Authenticate(ctx, addr, password)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrNotFound):
			return Token{}, errs.New(errs.FailedPrecondition, err)
		case errors.Is(err, user.ErrAuthenticationFailure):
			return Token{}, errs.New(errs.Unauthenticated, err)
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

	tkn, err := c.auth.GenerateToken(kid, claims)
	if err != nil {
		return Token{}, errs.New(errs.Internal, err)
	}

	return toToken(tkn), nil
}

// Create adds a new user to the system.
func (c *Core) Create(ctx context.Context, app NewUser) (User, error) {
	nc, err := toBusNewUser(app)
	if err != nil {
		return User{}, errs.New(errs.FailedPrecondition, err)
	}

	usr, err := c.user.Create(ctx, nc)
	if err != nil {
		if errors.Is(err, user.ErrUniqueEmail) {
			return User{}, errs.New(errs.Aborted, user.ErrUniqueEmail)
		}
		return User{}, errs.Newf(errs.Internal, "create: usr[%+v]: %s", usr, err)
	}

	return toAppUser(usr), nil
}

// Update updates an existing user.
func (c *Core) Update(ctx context.Context, app UpdateUser) (User, error) {
	uu, err := toBusUpdateUser(app)
	if err != nil {
		return User{}, errs.New(errs.FailedPrecondition, err)
	}

	usr, err := mid.GetUser(ctx)
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "user missing in context: %s", err)
	}

	updUsr, err := c.user.Update(ctx, usr, uu)
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "update: userID[%s] uu[%+v]: %s", usr.ID, uu, err)
	}

	return toAppUser(updUsr), nil
}

// UpdateRole updates an existing user's role.
func (c *Core) UpdateRole(ctx context.Context, app UpdateUserRole) (User, error) {
	uu, err := toBusUpdateUserRole(app)
	if err != nil {
		return User{}, errs.New(errs.FailedPrecondition, err)
	}

	usr, err := mid.GetUser(ctx)
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "user missing in context: %s", err)
	}

	updUsr, err := c.user.Update(ctx, usr, uu)
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "updaterole: userID[%s] uu[%+v]: %s", usr.ID, uu, err)
	}

	return toAppUser(updUsr), nil
}

// Delete removes a user from the system.
func (c *Core) Delete(ctx context.Context) error {
	usr, err := mid.GetUser(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "userID missing in context: %s", err)
	}

	if err := c.user.Delete(ctx, usr); err != nil {
		return errs.Newf(errs.Internal, "delete: userID[%s]: %s", usr.ID, err)
	}

	return nil
}

// Query returns a list of users with paging.
func (c *Core) Query(ctx context.Context, qp QueryParams) (page.Document[User], error) {
	if err := validatePaging(qp); err != nil {
		return page.Document[User]{}, err
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return page.Document[User]{}, err
	}

	orderBy, err := parseOrder(qp)
	if err != nil {
		return page.Document[User]{}, err
	}

	usrs, err := c.user.Query(ctx, filter, orderBy, qp.Page, qp.Rows)
	if err != nil {
		return page.Document[User]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := c.user.Count(ctx, filter)
	if err != nil {
		return page.Document[User]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return page.NewDocument(toAppUsers(usrs), total, qp.Page, qp.Rows), nil
}

// QueryByID returns a user by its ID.
func (c *Core) QueryByID(ctx context.Context) (User, error) {
	usr, err := mid.GetUser(ctx)
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return toAppUser(usr), nil
}
