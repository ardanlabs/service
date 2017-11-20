package handlers

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/internal/platform/web"
)

// User represents the User API method handler set.
type User struct {
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// List returns all the existing users in the system.
func (u *User) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {

	data := struct {
		Name  string
		Email string
	}{
		Name:  "Bill",
		Email: "bill@ardanlabs.com",
	}

	web.Respond(ctx, w, data, http.StatusOK)
	return nil
}
