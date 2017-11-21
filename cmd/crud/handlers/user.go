package handlers

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/internal/platform/db"
	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/ardanlabs/service/internal/user"
	"github.com/pkg/errors"
)

// User represents the User API method handler set.
type User struct {
	MasterDB *db.DB

	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// List returns all the existing users in the system.
func (u *User) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	dbConn, err := u.MasterDB.Copy()
	if err != nil {
		return errors.Wrapf(web.ErrDBNotConfigured, "")
	}
	defer dbConn.Close()

	usrs, err := user.List(ctx, dbConn)
	if err != nil {
		return errors.Wrap(err, "")
	}

	web.Respond(ctx, w, usrs, http.StatusOK)
	return nil
}

// Retrieve returns the specified user from the system.
func (u *User) Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	dbConn, err := u.MasterDB.Copy()
	if err != nil {
		return errors.Wrapf(web.ErrDBNotConfigured, "")
	}
	defer dbConn.Close()

	usr, err := user.Retrieve(ctx, dbConn, params["id"])
	if err != nil {
		return errors.Wrapf(err, "Id: %s", params["id"])
	}

	web.Respond(ctx, w, usr, http.StatusOK)
	return nil
}

// Create inserts a new user into the system.
func (u *User) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	dbConn, err := u.MasterDB.Copy()
	if err != nil {
		return errors.Wrapf(web.ErrDBNotConfigured, "")
	}
	defer dbConn.Close()

	var usr user.CreateUser
	if err := web.Unmarshal(r.Body, &usr); err != nil {
		return errors.Wrap(err, "")
	}

	nUsr, err := user.Create(ctx, dbConn, &usr)
	if err != nil {
		return errors.Wrapf(err, "User: %+v", &usr)
	}

	web.Respond(ctx, w, nUsr, http.StatusCreated)
	return nil
}

// Update updates the specified user in the system.
func (u *User) Update(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	dbConn, err := u.MasterDB.Copy()
	if err != nil {
		return errors.Wrapf(web.ErrDBNotConfigured, "")
	}
	defer dbConn.Close()

	var usr user.CreateUser
	if err := web.Unmarshal(r.Body, &usr); err != nil {
		return errors.Wrap(err, "")
	}

	if err := user.Update(ctx, dbConn, params["id"], &usr); err != nil {
		return errors.Wrapf(err, "Id: %s  User: %+v", params["id"], &usr)
	}

	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

// Delete removed the specified user from the system.
func (u *User) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	dbConn, err := u.MasterDB.Copy()
	if err != nil {
		return errors.Wrapf(web.ErrDBNotConfigured, "")
	}
	defer dbConn.Close()

	if err := user.Delete(ctx, dbConn, params["id"]); err != nil {
		return errors.Wrapf(err, "Id: %s", params["id"])
	}

	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}
