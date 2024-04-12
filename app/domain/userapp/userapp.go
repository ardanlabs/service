// Package userapp maintains the app layer api for the user domain.
package userapp

import (
	"context"
	"errors"

	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/app/api/mid"
	"github.com/ardanlabs/service/app/api/page"
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/domain/userbus"
)

// Core manages the set of app layer api functions for the user domain.
type Core struct {
	userBus *userbus.Core
	auth    *auth.Auth
}

// NewCore constructs a user core API for use.
func NewCore(userBus *userbus.Core) *Core {
	return &Core{
		userBus: userBus,
	}
}

// NewCoreWithAuth constructs a user core API for use with auth support.
func NewCoreWithAuth(userBus *userbus.Core, auth *auth.Auth) *Core {
	return &Core{
		auth:    auth,
		userBus: userBus,
	}
}

// Token provides an API token for the authenticated user.
func (c *Core) Token(ctx context.Context, kid string) (Token, error) {
	if c.auth == nil {
		return Token{}, errs.Newf(errs.Internal, "auth not configured")
	}

	claims := mid.GetClaims(ctx)

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

	usr, err := c.userBus.Create(ctx, nc)
	if err != nil {
		if errors.Is(err, userbus.ErrUniqueEmail) {
			return User{}, errs.New(errs.Aborted, userbus.ErrUniqueEmail)
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

	updUsr, err := c.userBus.Update(ctx, usr, uu)
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

	updUsr, err := c.userBus.Update(ctx, usr, uu)
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

	if err := c.userBus.Delete(ctx, usr); err != nil {
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

	usrs, err := c.userBus.Query(ctx, filter, orderBy, qp.Page, qp.Rows)
	if err != nil {
		return page.Document[User]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := c.userBus.Count(ctx, filter)
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
