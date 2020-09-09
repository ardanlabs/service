// Package handlers contains the full set of handler functions and routes
// supported by the web api.
package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/ardanlabs/service/business/auth" // Import is removed in final PR
	"github.com/ardanlabs/service/business/mid"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/jmoiron/sqlx"
)

// API constructs an http.Handler with all application routes defined.
func API(build string, shutdown chan os.Signal, log *log.Logger, db *sqlx.DB, a *auth.Auth) http.Handler {

	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	// Register health check endpoint. This route is not authenticated.
	c := check{
		build: build,
		db:    db,
	}
	app.Handle(http.MethodGet, "/v1/readiness", c.readiness)
	app.Handle(http.MethodGet, "/v1/liveness", c.liveness)

	// Register user management and authentication endpoints.
	u := userHandlers{
		db:   db,
		auth: a,
	}
	app.Handle(http.MethodGet, "/v1/users", u.query, mid.Authenticate(a), mid.HasRole(auth.RoleAdmin))
	app.Handle(http.MethodPost, "/v1/users", u.create, mid.Authenticate(a), mid.HasRole(auth.RoleAdmin))
	app.Handle(http.MethodGet, "/v1/users/:id", u.queryByID, mid.Authenticate(a))
	app.Handle(http.MethodPut, "/v1/users/:id", u.update, mid.Authenticate(a), mid.HasRole(auth.RoleAdmin))
	app.Handle(http.MethodDelete, "/v1/users/:id", u.delete, mid.Authenticate(a), mid.HasRole(auth.RoleAdmin))

	// This route is not authenticated.
	app.Handle(http.MethodGet, "/v1/users/token", u.token)

	// Register product and sale endpoints.
	p := productHandlers{
		db: db,
	}
	app.Handle(http.MethodGet, "/v1/products", p.query, mid.Authenticate(a))
	app.Handle(http.MethodPost, "/v1/products", p.create, mid.Authenticate(a))
	app.Handle(http.MethodGet, "/v1/products/:id", p.queryByID, mid.Authenticate(a))
	app.Handle(http.MethodPut, "/v1/products/:id", p.update, mid.Authenticate(a))
	app.Handle(http.MethodDelete, "/v1/products/:id", p.delete, mid.Authenticate(a))

	return app
}
