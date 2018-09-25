package handlers

import (
	"context"
	"crypto/rsa"
	"log"
	"net/http"
	"time"

	"github.com/ardanlabs/service/internal/platform/db"
	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/ardanlabs/service/internal/user"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

// UserAuth knows everything needed to authenticate Users.
type UserAuth struct {
	Key   *rsa.PrivateKey
	KeyID string
	Alg   string
}

// User represents the User API method handler set.
type User struct {
	MasterDB *db.DB
	Auth     UserAuth

	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// List returns all the existing users in the system.
func (u *User) List(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.List")
	defer span.End()

	dbConn := u.MasterDB.Copy()
	defer dbConn.Close()

	usrs, err := user.List(ctx, dbConn)
	if err = translate(err); err != nil {
		return errors.Wrap(err, "")
	}

	web.Respond(ctx, log, w, usrs, http.StatusOK)
	return nil
}

// Retrieve returns the specified user from the system.
func (u *User) Retrieve(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.Retrieve")
	defer span.End()

	dbConn := u.MasterDB.Copy()
	defer dbConn.Close()

	usr, err := user.Retrieve(ctx, dbConn, params["id"])
	if err = translate(err); err != nil {
		return errors.Wrapf(err, "Id: %s", params["id"])
	}

	web.Respond(ctx, log, w, usr, http.StatusOK)
	return nil
}

// Create inserts a new user into the system.
func (u *User) Create(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.Create")
	defer span.End()

	dbConn := u.MasterDB.Copy()
	defer dbConn.Close()

	var newU user.NewUser
	if err := web.Unmarshal(r.Body, &newU); err != nil {
		return errors.Wrap(err, "")
	}

	// TODO(jlw) use time from request context
	usr, err := user.Create(ctx, dbConn, &newU, time.Now().UTC())
	if err = translate(err); err != nil {
		return errors.Wrapf(err, "User: %+v", &usr)
	}

	web.Respond(ctx, log, w, usr, http.StatusCreated)
	return nil
}

// Update updates the specified user in the system.
func (u *User) Update(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.Update")
	defer span.End()

	dbConn := u.MasterDB.Copy()
	defer dbConn.Close()

	var upd user.UpdateUser
	if err := web.Unmarshal(r.Body, &upd); err != nil {
		return errors.Wrap(err, "")
	}

	// TODO(jlw) use time from request context
	err := user.Update(ctx, dbConn, params["id"], &upd, time.Now().UTC())
	if err = translate(err); err != nil {
		return errors.Wrapf(err, "Id: %s  User: %+v", params["id"], &upd)
	}

	web.Respond(ctx, log, w, nil, http.StatusNoContent)
	return nil
}

// Delete removed the specified user from the system.
func (u *User) Delete(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.Delete")
	defer span.End()

	dbConn := u.MasterDB.Copy()
	defer dbConn.Close()

	err := user.Delete(ctx, dbConn, params["id"])
	if err = translate(err); err != nil {
		return errors.Wrapf(err, "Id: %s", params["id"])
	}

	web.Respond(ctx, log, w, nil, http.StatusNoContent)
	return nil
}

// Token handles a request to authenticate a user. It expects a request using
// Basic Auth with a user's email and password. It responds with a JWT.
func (u *User) Token(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.Token")
	defer span.End()

	dbConn := u.MasterDB.Copy()
	defer dbConn.Close()

	email, pass, ok := r.BasicAuth()
	if !ok {
		return web.ErrUnauthorized
	}

	tkn, err := user.Authenticate(ctx, dbConn, u.Auth.Key, u.Auth.KeyID, u.Auth.Alg, email, pass)
	if err = translate(err); err != nil {
		return errors.Wrap(err, "authenticating")
	}

	web.Respond(ctx, log, w, tkn, http.StatusOK)
	return nil
}
