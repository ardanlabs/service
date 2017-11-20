package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/internal/platform/db"
	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/pkg/errors"
	mgo "gopkg.in/mgo.v2"
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

	data := struct {
		Name  string
		Email string
	}{
		Name:  "Bill",
		Email: "bill@ardanlabs.com",
	}

	f := func(collection *mgo.Collection) error {
		return collection.Insert(data)
	}
	if err := dbConn.Execute(ctx, "users", f); err != nil {
		return errors.Wrap(err, fmt.Sprintf("db.users.insert(%s)", db.Query(u)))
	}

	web.Respond(ctx, w, data, http.StatusCreated)
	return nil
}
