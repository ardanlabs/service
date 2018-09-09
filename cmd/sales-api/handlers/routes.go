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

	// Register health check endpoint.
	h := Health{
		MasterDB: masterDB,
	}
	app.Handle("GET", "/v1/health", h.Check)

	u := User{
		MasterDB: masterDB,
	}
	app.Handle("GET", "/v1/users", u.List)
	app.Handle("POST", "/v1/users", u.Create)
	app.Handle("GET", "/v1/users/:id", u.Retrieve)
	app.Handle("PUT", "/v1/users/:id", u.Update)
	app.Handle("DELETE", "/v1/users/:id", u.Delete)

	// Register product and sale endpoints.
	p := Product{
		MasterDB: masterDB,
	}
	app.Handle("GET", "/v1/products", p.List)
	app.Handle("POST", "/v1/products", p.Create)
	app.Handle("GET", "/v1/products/:id", p.Retrieve)
	app.Handle("PUT", "/v1/products/:id", p.Update)
	app.Handle("DELETE", "/v1/products/:id", p.Delete)

	return app
}
