package vproductgrp

import (
	"net/http"

	"github.com/ardanlabs/service/app/services/sales-api/apis/views/vproductapi"
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/api/mid"
	"github.com/ardanlabs/service/business/core/views/vproduct"
	"github.com/ardanlabs/service/business/core/views/vproduct/stores/vproductdb"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/jmoiron/sqlx"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log  *logger.Logger
	Auth *auth.Auth
	DB   *sqlx.DB
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	vproductCore := vproduct.NewCore(vproductdb.NewStore(cfg.Log, cfg.DB))

	authen := mid.Authenticate(cfg.Auth)
	ruleAdmin := mid.Authorize(cfg.Auth, auth.RuleAdminOnly)

	hdl := new(vproductapi.New(vproductCore))
	app.Handle(http.MethodGet, version, "/vproducts", hdl.query, authen, ruleAdmin)
}
