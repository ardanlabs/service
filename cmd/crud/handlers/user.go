package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/ardanlabs/service/internal/platform/db"
	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/ardanlabs/service/internal/user"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

// check looks for certain error types and transforms them
// into web errors. We are losing the trace when this error
// is converted. But we don't log traces for these.
func check(err error) error {
	switch errors.Cause(err) {
	case user.ErrNotFound:
		return web.ErrNotFound
	case user.ErrInvalidID:
		return web.ErrInvalidID
	}
	return err
}

// =============================================================================

// User represents the User API method handler set.
type User struct {
	MasterDB *db.DB

	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// List returns all the existing users in the system.
func (u *User) List(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.List")
	defer span.End()

	dbConn := u.MasterDB.Copy()
	defer dbConn.Close()

	usrs, err := user.List(ctx, dbConn)
	if err = check(err); err != nil {
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
	if err = check(err); err != nil {
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

	var usr user.CreateUser
	if err := web.Unmarshal(r.Body, &usr); err != nil {
		return errors.Wrap(err, "")
	}

	nUsr, err := user.Create(ctx, dbConn, &usr, time.Now().UTC())
	if err = check(err); err != nil {
		return errors.Wrapf(err, "User: %+v", &usr)
	}

	web.Respond(ctx, log, w, nUsr, http.StatusCreated)
	return nil
}

// Update updates the specified user in the system.
func (u *User) Update(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.User.Update")
	defer span.End()

	dbConn := u.MasterDB.Copy()
	defer dbConn.Close()

	var usr user.CreateUser
	if err := web.Unmarshal(r.Body, &usr); err != nil {
		return errors.Wrap(err, "")
	}

	err := user.Update(ctx, dbConn, params["id"], &usr, time.Now().UTC())
	if err = check(err); err != nil {
		return errors.Wrapf(err, "Id: %s  User: %+v", params["id"], &usr)
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
	if err = check(err); err != nil {
		return errors.Wrapf(err, "Id: %s", params["id"])
	}

	web.Respond(ctx, log, w, nil, http.StatusNoContent)
	return nil
}
