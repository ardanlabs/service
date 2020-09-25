// Package handlers contains the full set of handler functions and routes
// supported by the web api.
package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/ardanlabs/service/business/auth" // Import is removed in final PR
	"github.com/ardanlabs/service/business/data/product"
	"github.com/ardanlabs/service/business/data/user"
	"github.com/ardanlabs/service/business/mid"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/jmoiron/sqlx"
)

// API constructs an http.Handler with all application routes defined.
func API(build string, shutdown chan os.Signal, log *log.Logger, db *sqlx.DB, a *auth.Auth) http.Handler {

	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	// Register health check endpoint. This route is not authenticated.
	cg := checkGroup{
		build: build,
		db:    db,
	}
	app.Handle(http.MethodGet, "/v1/readiness", cg.readiness)
	app.Handle(http.MethodGet, "/v1/liveness", cg.liveness)

	// Register user management and authentication endpoints.
	ug := userGroup{
		user: user.New(log, db),
		auth: a,
	}
	app.Handle(http.MethodGet, "/v1/users", ug.query, mid.Authenticate(a), mid.HasRole(auth.RoleAdmin))
	app.Handle(http.MethodPost, "/v1/users", ug.create, mid.Authenticate(a), mid.HasRole(auth.RoleAdmin))
	app.Handle(http.MethodGet, "/v1/users/:id", ug.queryByID, mid.Authenticate(a))
	app.Handle(http.MethodPut, "/v1/users/:id", ug.update, mid.Authenticate(a), mid.HasRole(auth.RoleAdmin))
	app.Handle(http.MethodDelete, "/v1/users/:id", ug.delete, mid.Authenticate(a), mid.HasRole(auth.RoleAdmin))
	app.Handle(http.MethodGet, "/v1/users/token/:kid", ug.token)

	// Register product and sale endpoints.
	pg := productGroup{
		product: product.New(log, db),
	}
	app.Handle(http.MethodGet, "/v1/products", pg.query, mid.Authenticate(a))
	app.Handle(http.MethodPost, "/v1/products", pg.create, mid.Authenticate(a))
	app.Handle(http.MethodGet, "/v1/products/:id", pg.queryByID, mid.Authenticate(a))
	app.Handle(http.MethodPut, "/v1/products/:id", pg.update, mid.Authenticate(a))
	app.Handle(http.MethodDelete, "/v1/products/:id", pg.delete, mid.Authenticate(a))

	return app
}
