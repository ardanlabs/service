package handlers

import (
	"log"
	"net/http"

	"github.com/ardanlabs/service/internal/mid"
	"github.com/ardanlabs/service/internal/platform/db"
	"github.com/ardanlabs/service/internal/platform/web"
)

// API returns a handler for a set of routes.
func API(log *log.Logger, masterDB *db.DB) http.Handler {
	app := web.New(log, mid.RequestLogger, mid.Metrics, mid.ErrorHandler)

	u := User{
		Log:      log,
		MasterDB: masterDB,
	}
	app.Handle("GET", "/v1/users", u.List)
	app.Handle("POST", "/v1/users", u.Create)
	app.Handle("GET", "/v1/users/:id", u.Retrieve)
	app.Handle("PUT", "/v1/users/:id", u.Update)
	app.Handle("DELETE", "/v1/users/:id", u.Delete)

	h := Health{
		Log:      log,
		MasterDB: masterDB,
	}
	app.Handle("GET", "/v1/health", h.Check)

	return app
}
