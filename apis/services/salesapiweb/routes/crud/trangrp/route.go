package trangrp

import (
	"net/http"

	"github.com/ardanlabs/service/app/core/crud/tranapp"
	"github.com/ardanlabs/service/business/api/auth"
	midhttp "github.com/ardanlabs/service/business/api/mid/http"
	"github.com/ardanlabs/service/business/core/crud/product"
	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/data/sqldb"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/jmoiron/sqlx"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	UserCore    *user.Core
	ProductCore *product.Core
	Log         *logger.Logger
	DB          *sqlx.DB
	Auth        *auth.Auth
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := midhttp.Authenticate(cfg.Auth)
	tran := midhttp.ExecuteInTransaction(cfg.Log, sqldb.NewBeginner(cfg.DB))

	hdl := new(tranapp.New(cfg.UserCore, cfg.ProductCore))
	app.Handle(http.MethodPost, version, "/tranexample", hdl.create, authen, tran)
}
