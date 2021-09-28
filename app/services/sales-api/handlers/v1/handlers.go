// Package v1 contains the full set of handler functions and routes
// supported by the v1 web api.
package v1

import (
	"net/http"

	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/productgrp"
	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/usergrp"
	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/sys/auth"
	"github.com/ardanlabs/service/business/web/v1/mid"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log  *zap.SugaredLogger
	Auth *auth.Auth
	DB   *sqlx.DB
}

// Routes binds all the version 1 routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	// Register user management and authentication endpoints.
	ugh := usergrp.Handlers{
		Core: user.NewCore(cfg.Log, cfg.DB),
		Auth: cfg.Auth,
	}
	app.Handle(http.MethodGet, version, "/users/token", ugh.Token)
	app.Handle(http.MethodGet, version, "/users/:page/:rows", ugh.Query, mid.Authenticate(cfg.Auth), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodGet, version, "/users/:id", ugh.QueryByID, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodPost, version, "/users", ugh.Create, mid.Authenticate(cfg.Auth), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodPut, version, "/users/:id", ugh.Update, mid.Authenticate(cfg.Auth), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodDelete, version, "/users/:id", ugh.Delete, mid.Authenticate(cfg.Auth), mid.Authorize(auth.RoleAdmin))

	// Register product and sale endpoints.
	pgh := productgrp.Handlers{
		Core: product.NewCore(cfg.Log, cfg.DB),
	}
	app.Handle(http.MethodGet, version, "/products/:page/:rows", pgh.Query, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodGet, version, "/products/:id", pgh.QueryByID, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodPost, version, "/products", pgh.Create, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodPut, version, "/products/:id", pgh.Update, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodDelete, version, "/products/:id", pgh.Delete, mid.Authenticate(cfg.Auth))
}
