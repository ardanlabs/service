package productapi

import (
	"net/http"

	"github.com/ardanlabs/service/apis/services/sales/mid"
	"github.com/ardanlabs/service/app/api/authsrv"
	"github.com/ardanlabs/service/app/core/crud/productapp"
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/core/crud/productbus"
	"github.com/ardanlabs/service/business/core/crud/userbus"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log        *logger.Logger
	UserBus    *userbus.Core
	ProductBus *productbus.Core
	AuthSrv    *authsrv.AuthSrv
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.Log, cfg.AuthSrv)
	ruleAny := mid.Authorize(cfg.Log, cfg.AuthSrv, auth.RuleAny)
	ruleUserOnly := mid.Authorize(cfg.Log, cfg.AuthSrv, auth.RuleUserOnly)
	ruleAuthorizeProduct := mid.AuthorizeProduct(cfg.Log, cfg.AuthSrv, cfg.ProductBus)

	api := newAPI(productapp.NewCore(cfg.ProductBus))
	app.Handle(http.MethodGet, version, "/products", api.query, authen, ruleAny)
	app.Handle(http.MethodGet, version, "/products/{product_id}", api.queryByID, authen, ruleAuthorizeProduct)
	app.Handle(http.MethodPost, version, "/products", api.create, authen, ruleUserOnly)
	app.Handle(http.MethodPut, version, "/products/{product_id}", api.update, authen, ruleAuthorizeProduct)
	app.Handle(http.MethodDelete, version, "/products/{product_id}", api.delete, authen, ruleAuthorizeProduct)
}
