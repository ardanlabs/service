package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/ardanlabs/service/internal/mid"
	"github.com/ardanlabs/service/internal/platform/auth"
	"github.com/ardanlabs/service/internal/platform/db"
	"github.com/ardanlabs/service/internal/platform/web"
)

// API returns a handler for a set of routes.
func API(shutdown chan os.Signal, log *log.Logger, masterDB *db.DB, authenticator *auth.Authenticator) http.Handler {

	// Create the variable that contains all Middleware functions.
	mw := mid.Middleware{Authenticator: authenticator}

	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.New(shutdown, log, mw.Logger, mw.Errors, mw.Metrics, mw.Panics)

	// Register health check endpoint. This route is not authenticated.
	check := Check{
		MasterDB: masterDB,
	}
	app.Handle("GET", "/v1/health", check.Health)

	// Register user management and authentication endpoints.
	u := User{
		MasterDB:       masterDB,
		TokenGenerator: authenticator,
	}
	app.Handle("GET", "/v1/users", u.List, mw.Authenticate, mw.HasRole(auth.RoleAdmin))
	app.Handle("POST", "/v1/users", u.Create, mw.Authenticate, mw.HasRole(auth.RoleAdmin))
	app.Handle("GET", "/v1/users/:id", u.Retrieve, mw.Authenticate)
	app.Handle("PUT", "/v1/users/:id", u.Update, mw.Authenticate, mw.HasRole(auth.RoleAdmin))
	app.Handle("DELETE", "/v1/users/:id", u.Delete, mw.Authenticate, mw.HasRole(auth.RoleAdmin))

	// This route is not authenticated
	app.Handle("GET", "/v1/users/token", u.Token)

	// Register product and sale endpoints.
	p := Product{
		MasterDB: masterDB,
	}
	app.Handle("GET", "/v1/products", p.List, mw.Authenticate)
	app.Handle("POST", "/v1/products", p.Create, mw.Authenticate)
	app.Handle("GET", "/v1/products/:id", p.Retrieve, mw.Authenticate)
	app.Handle("PUT", "/v1/products/:id", p.Update, mw.Authenticate)
	app.Handle("DELETE", "/v1/products/:id", p.Delete, mw.Authenticate)

	return app
}
