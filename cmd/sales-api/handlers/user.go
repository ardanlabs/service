package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/ardanlabs/service/internal/platform/auth"
	"github.com/ardanlabs/service/internal/platform/db"
	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/ardanlabs/service/internal/user"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

// User represents the User API method handler set.
type User struct {
	MasterDB       *db.DB
	TokenGenerator user.TokenGenerator

	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// List returns all the existing users in the system.
func (u *User) List(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.List")
	defer span.End()

	dbConn := u.MasterDB.Copy()
	defer dbConn.Close()

	usrs, err := user.List(ctx, dbConn)
	if err != nil {
		return err
	}

	return web.Respond(ctx, log, w, usrs, http.StatusOK)
}

// Retrieve returns the specified user from the system.
func (u *User) Retrieve(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.Retrieve")
	defer span.End()

	dbConn := u.MasterDB.Copy()
	defer dbConn.Close()

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	usr, err := user.Retrieve(ctx, claims, dbConn, params["id"])
	if err != nil {
		switch err {
		case user.ErrInvalidID:
			return web.WrapErrorWithStatus(err, http.StatusBadRequest)
		case user.ErrNotFound:
			return web.WrapErrorWithStatus(err, http.StatusNotFound)
		case user.ErrForbidden:
			return web.WrapErrorWithStatus(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "Id: %s", params["id"])
		}
	}

	return web.Respond(ctx, log, w, usr, http.StatusOK)
}

// Create inserts a new user into the system.
func (u *User) Create(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.Create")
	defer span.End()

	dbConn := u.MasterDB.Copy()
	defer dbConn.Close()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.Shutdown("web value missing from context")
	}

	var newU user.NewUser
	if err := web.Decode(r, &newU); err != nil {
		return errors.Wrap(err, "")
	}

	usr, err := user.Create(ctx, dbConn, &newU, v.Now)
	if err != nil {
		return errors.Wrapf(err, "User: %+v", &usr)
	}

	return web.Respond(ctx, log, w, usr, http.StatusCreated)
}

// Update updates the specified user in the system.
func (u *User) Update(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.Update")
	defer span.End()

	dbConn := u.MasterDB.Copy()
	defer dbConn.Close()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.Shutdown("web value missing from context")
	}

	var upd user.UpdateUser
	if err := web.Decode(r, &upd); err != nil {
		return errors.Wrap(err, "")
	}

	err := user.Update(ctx, dbConn, params["id"], &upd, v.Now)
	if err != nil {
		switch err {
		case user.ErrInvalidID:
			return web.WrapErrorWithStatus(err, http.StatusBadRequest)
		case user.ErrNotFound:
			return web.WrapErrorWithStatus(err, http.StatusNotFound)
		case user.ErrForbidden:
			return web.WrapErrorWithStatus(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "Id: %s  User: %+v", params["id"], &upd)
		}
	}

	return web.Respond(ctx, log, w, nil, http.StatusNoContent)
}

// Delete removes the specified user from the system.
func (u *User) Delete(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.Delete")
	defer span.End()

	dbConn := u.MasterDB.Copy()
	defer dbConn.Close()

	err := user.Delete(ctx, dbConn, params["id"])
	if err != nil {
		switch err {
		case user.ErrInvalidID:
			return web.WrapErrorWithStatus(err, http.StatusBadRequest)
		case user.ErrNotFound:
			return web.WrapErrorWithStatus(err, http.StatusNotFound)
		case user.ErrForbidden:
			return web.WrapErrorWithStatus(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "Id: %s", params["id"])
		}
	}

	return web.Respond(ctx, log, w, nil, http.StatusNoContent)
}

// Token handles a request to authenticate a user. It expects a request using
// Basic Auth with a user's email and password. It responds with a JWT.
func (u *User) Token(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.Token")
	defer span.End()

	dbConn := u.MasterDB.Copy()
	defer dbConn.Close()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.Shutdown("web value missing from context")
	}

	email, pass, ok := r.BasicAuth()
	if !ok {
		err := errors.New("must provide email and password in Basic auth")
		return web.WrapErrorWithStatus(err, http.StatusUnauthorized)
	}

	tkn, err := user.Authenticate(ctx, dbConn, u.TokenGenerator, v.Now, email, pass)
	if err != nil {
		switch err {
		case user.ErrAuthenticationFailure:
			return web.WrapErrorWithStatus(err, http.StatusUnauthorized)
		default:
			return errors.Wrap(err, "authenticating")
		}
	}

	return web.Respond(ctx, log, w, tkn, http.StatusOK)
}
