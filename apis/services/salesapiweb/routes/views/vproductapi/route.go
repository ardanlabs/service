package vproductapi

import (
	"net/http"

	midhttp "github.com/ardanlabs/service/app/api/mid/http"
	"github.com/ardanlabs/service/app/core/views/vproductapp"
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/core/views/vproductbus"
	"github.com/ardanlabs/service/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	VProductBus *vproductbus.Core
	Auth        *auth.Auth
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := midhttp.Authenticate(cfg.Auth)
	ruleAdmin := midhttp.Authorize(cfg.Auth, auth.RuleAdminOnly)

	api := newAPI(vproductapp.NewCore(cfg.VProductBus))
	app.Handle(http.MethodGet, version, "/vproducts", api.query, authen, ruleAdmin)
}
