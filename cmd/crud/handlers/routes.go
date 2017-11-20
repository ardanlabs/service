package handlers

import (
	"net/http"

	"github.com/ardanlabs/service/internal/mid"
	"github.com/ardanlabs/service/internal/platform/db"
	"github.com/ardanlabs/service/internal/platform/web"
)

// API returns a handler for a set of routes.
func API(masterDB *db.DB) http.Handler {

	// Create the application.
	app := web.New(mid.RequestLogger, mid.ErrorHandler)

	// Bind all the user handlers.
	u := User{
		MasterDB: masterDB,
	}
	app.Handle("GET", "/v1/users", u.List)

	return app
}
