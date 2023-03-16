// Package v1 contains the full set of handler functions and routes
// supported by the v1 web api.
package v1

import (
	"net/http"

	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/productgrp"
	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/usergrp"
	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/product/stores/productdb"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/core/user/stores/usercache"
	"github.com/ardanlabs/service/business/core/user/stores/userdb"
	"github.com/ardanlabs/service/business/web/auth"
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

	usrCore := user.NewCore(usercache.NewStore(cfg.Log, userdb.NewStore(cfg.Log, cfg.DB)))
	prdCore := product.NewCore(usrCore, productdb.NewStore(cfg.Log, cfg.DB))

	authen := mid.Authenticate(cfg.Auth)
	ruleAdmin := mid.Authorize(cfg.Auth, auth.RuleAdminOnly)
	ruleAdminOrSubject := mid.Authorize(cfg.Auth, auth.RuleAdminOrSubject)

	ugh := usergrp.Handlers{
		User: usrCore,
		Auth: cfg.Auth,
	}
	app.Handle(http.MethodGet, version, "/users/token/:kid", ugh.Token)
	app.Handle(http.MethodGet, version, "/users", ugh.Query, authen, ruleAdmin)
	app.Handle(http.MethodGet, version, "/users/:user_id", ugh.QueryByID, authen, ruleAdminOrSubject)
	app.Handle(http.MethodPost, version, "/users", ugh.Create, authen, ruleAdmin)
	app.Handle(http.MethodPut, version, "/users/:user_id", ugh.Update, authen, ruleAdminOrSubject)
	app.Handle(http.MethodDelete, version, "/users/:user_id", ugh.Delete, authen, ruleAdminOrSubject)

	pgh := productgrp.Handlers{
		Product: prdCore,
		Auth:    cfg.Auth,
	}
	app.Handle(http.MethodGet, version, "/products", pgh.Query, authen)
	app.Handle(http.MethodGet, version, "/products/:product_id", pgh.QueryByID, authen)
	app.Handle(http.MethodPost, version, "/products", pgh.Create, authen)
	app.Handle(http.MethodPut, version, "/products/:product_id", pgh.Update, authen)
	app.Handle(http.MethodDelete, version, "/products/:product_id", pgh.Delete, authen)
}
