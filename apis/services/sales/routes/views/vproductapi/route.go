package vproductapi

import (
	"net/http"

	"github.com/ardanlabs/service/apis/services/sales/mid"
	"github.com/ardanlabs/service/app/api/authsrv"
	"github.com/ardanlabs/service/app/core/views/vproductapp"
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/core/crud/userbus"
	"github.com/ardanlabs/service/business/core/views/vproductbus"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log         *logger.Logger
	UserBus     *userbus.Core
	VProductBus *vproductbus.Core
	AuthSrv     *authsrv.AuthSrv
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.Log, cfg.AuthSrv)
	ruleAdmin := mid.Authorize(cfg.Log, cfg.AuthSrv, auth.RuleAdminOnly)

	api := newAPI(vproductapp.NewCore(cfg.VProductBus))
	app.Handle(http.MethodGet, version, "/vproducts", api.query, authen, ruleAdmin)
}
