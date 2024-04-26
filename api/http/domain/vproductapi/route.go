package vproductapi

import (
	"net/http"

	"github.com/ardanlabs/service/api/http/api/mid"
	"github.com/ardanlabs/service/app/api/auth"
	"github.com/ardanlabs/service/app/api/authclient"
	"github.com/ardanlabs/service/app/domain/vproductapp"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/business/domain/vproductbus"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log         *logger.Logger
	UserBus     *userbus.Core
	VProductBus *vproductbus.Core
	AuthClient  *authclient.Client
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.Log, cfg.AuthClient)
	ruleAdmin := mid.Authorize(cfg.Log, cfg.AuthClient, auth.RuleAdminOnly)

	api := newAPI(vproductapp.NewCore(cfg.VProductBus))
	app.Handle(http.MethodGet, version, "/vproducts", api.query, authen, ruleAdmin)
}
