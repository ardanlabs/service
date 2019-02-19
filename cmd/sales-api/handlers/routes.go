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
	app := web.New(shutdown, log, mid.ErrorHandler, mid.Metrics, mid.RequestLogger)

	// authmw is used for authentication/authorization middleware.
	authmw := mid.Auth{
		Authenticator: authenticator,
	}

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
	app.Handle("GET", "/v1/users", u.List, authmw.HasRole(auth.RoleAdmin), authmw.Authenticate)
	app.Handle("POST", "/v1/users", u.Create, authmw.HasRole(auth.RoleAdmin), authmw.Authenticate)
	app.Handle("GET", "/v1/users/:id", u.Retrieve, authmw.Authenticate)
	app.Handle("PUT", "/v1/users/:id", u.Update, authmw.HasRole(auth.RoleAdmin), authmw.Authenticate)
	app.Handle("DELETE", "/v1/users/:id", u.Delete, authmw.HasRole(auth.RoleAdmin), authmw.Authenticate)

	// This route is not authenticated
	app.Handle("GET", "/v1/users/token", u.Token)

	// Register product and sale endpoints.
	p := Product{
		MasterDB: masterDB,
	}
	app.Handle("GET", "/v1/products", p.List, authmw.Authenticate)
	app.Handle("POST", "/v1/products", p.Create, authmw.Authenticate)
	app.Handle("GET", "/v1/products/:id", p.Retrieve, authmw.Authenticate)
	app.Handle("PUT", "/v1/products/:id", p.Update, authmw.Authenticate)
	app.Handle("DELETE", "/v1/products/:id", p.Delete, authmw.Authenticate)

	return app
}
