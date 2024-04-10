package productapi

import (
	"net/http"

	midhttp "github.com/ardanlabs/service/app/api/mid/http"
	"github.com/ardanlabs/service/app/core/crud/productapp"
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/core/crud/productbus"
	"github.com/ardanlabs/service/business/core/crud/userbus"
	"github.com/ardanlabs/service/foundation/authapi"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log        *logger.Logger
	UserBus    *userbus.Core
	ProductBus *productbus.Core
	AuthAPI    *authapi.AuthAPI
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := midhttp.AuthenticateWeb(cfg.Log, cfg.AuthAPI)
	ruleAny := midhttp.Authorize(cfg.Log, cfg.AuthAPI, auth.RuleAny)
	ruleUserOnly := midhttp.Authorize(cfg.Log, cfg.AuthAPI, auth.RuleUserOnly)
	ruleAuthorizeProduct := midhttp.AuthorizeProduct(cfg.Log, cfg.AuthAPI, cfg.ProductBus)

	api := newAPI(productapp.NewCore(cfg.ProductBus))
	app.Handle(http.MethodGet, version, "/products", api.query, authen, ruleAny)
	app.Handle(http.MethodGet, version, "/products/{product_id}", api.queryByID, authen, ruleAuthorizeProduct)
	app.Handle(http.MethodPost, version, "/products", api.create, authen, ruleUserOnly)
	app.Handle(http.MethodPut, version, "/products/{product_id}", api.update, authen, ruleAuthorizeProduct)
	app.Handle(http.MethodDelete, version, "/products/{product_id}", api.delete, authen, ruleAuthorizeProduct)
}
