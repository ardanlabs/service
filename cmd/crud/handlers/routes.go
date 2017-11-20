package handlers

import (
	"net/http"

	"github.com/ardanlabs/service/internal/mid"

	"github.com/ardanlabs/service/internal/platform/web"
)

// API returns a handler for a set of routes.
func API() http.Handler {

	// Create the application.
	app := web.New(mid.RequestLogger, mid.ErrorHandler)

	// Bind all the user handlers.
	var u User
	app.Handle("GET", "/v1/users", u.List)

	return app
}
