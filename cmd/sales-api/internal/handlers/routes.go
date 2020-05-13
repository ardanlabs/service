package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/ardanlabs/service/internal/mid"
	"github.com/ardanlabs/service/internal/platform/auth" // Import is removed in final PR
	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/jmoiron/sqlx"
)

// API constructs an http.Handler with all application routes defined.
func API(build string, shutdown chan os.Signal, log *log.Logger, db *sqlx.DB, authenticator *auth.Authenticator) http.Handler {

	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	// Register health check endpoint. This route is not authenticated.
	c := check{
		build: build,
		db:    db,
	}
	app.Handle(http.MethodGet, "/v1/health", c.health)

	// Register user management and authentication endpoints.
	u := user{
		db:            db,
		authenticator: authenticator,
	}
	app.Handle(http.MethodGet, "/v1/users", u.list, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle(http.MethodPost, "/v1/users", u.create, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle(http.MethodGet, "/v1/users/:id", u.retrieve, mid.Authenticate(authenticator))
	app.Handle(http.MethodPut, "/v1/users/:id", u.update, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))
	app.Handle(http.MethodDelete, "/v1/users/:id", u.delete, mid.Authenticate(authenticator), mid.HasRole(auth.RoleAdmin))

	// This route is not authenticated
	app.Handle(http.MethodGet, "/v1/users/token", u.token)

	// Register product and sale endpoints.
	p := product{
		db: db,
	}
	app.Handle(http.MethodGet, "/v1/products", p.list, mid.Authenticate(authenticator))
	app.Handle(http.MethodPost, "/v1/products", p.create, mid.Authenticate(authenticator))
	app.Handle(http.MethodGet, "/v1/products/:id", p.retrieve, mid.Authenticate(authenticator))
	app.Handle(http.MethodPut, "/v1/products/:id", p.update, mid.Authenticate(authenticator))
	app.Handle(http.MethodDelete, "/v1/products/:id", p.delete, mid.Authenticate(authenticator))

	return app
}
