package trangrp

import (
	"net/http"

	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/mid"
	"github.com/ardanlabs/service/business/core/event"
	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/product/stores/productdb"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/core/user/stores/usercache"
	"github.com/ardanlabs/service/business/core/user/stores/userdb"
	db "github.com/ardanlabs/service/business/data/dbsql/pq"
	"github.com/ardanlabs/service/business/web/auth"
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

	envCore := event.NewCore(cfg.Log)
	usrCore := user.NewCore(cfg.Log, envCore, usercache.NewStore(cfg.Log, userdb.NewStore(cfg.Log, cfg.DB)))
	prdCore := product.NewCore(cfg.Log, envCore, usrCore, productdb.NewStore(cfg.Log, cfg.DB))

	authen := mid.Authenticate(cfg.Auth)
	tran := mid.ExecuteInTransation(cfg.Log, db.NewBeginner(cfg.DB))

	hdl := New(usrCore, prdCore)
	app.Handle(http.MethodPost, version, "/tranexample", hdl.Create, authen, tran)
}
