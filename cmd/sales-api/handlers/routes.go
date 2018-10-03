package handlers

import (
	"log"
	"net/http"

	"github.com/ardanlabs/service/internal/mid"
	"github.com/ardanlabs/service/internal/platform/auth"
	"github.com/ardanlabs/service/internal/platform/db"
	"github.com/ardanlabs/service/internal/platform/web"
)

// API returns a handler for a set of routes.
func API(log *log.Logger, masterDB *db.DB, authenticator *auth.Authenticator) http.Handler {

	// authmw is used for authentication/authorization middleware.
	authmw := mid.Auth{
		Authenticator: authenticator,
	}

	// TODO(jlw) Figure out why the order of these was reversed and maybe reverse it back.
	app := web.New(log, mid.RequestLogger, mid.Metrics, mid.ErrorHandler)

	// Register health check endpoint. This route is not authenticated.
	h := Health{
		MasterDB: masterDB,
	}
	app.Handle("GET", "/v1/health", h.Check)

	// Register user management and authentication endpoints.
	u := User{
		MasterDB:      masterDB,
		Authenticator: authenticator,
	}

	app.Handle("GET", "/v1/users", u.List, authmw.Authenticate, authmw.HasRole(auth.RoleAdmin))
	app.Handle("POST", "/v1/users", u.Create, authmw.Authenticate, authmw.HasRole(auth.RoleAdmin))
	app.Handle("GET", "/v1/users/:id", u.Retrieve, authmw.Authenticate)
	app.Handle("PUT", "/v1/users/:id", u.Update, authmw.Authenticate, authmw.HasRole(auth.RoleAdmin))
	app.Handle("DELETE", "/v1/users/:id", u.Delete, authmw.Authenticate, authmw.HasRole(auth.RoleAdmin))

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
