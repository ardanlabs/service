package productapi

import (
	"net/http"

	"github.com/ardanlabs/service/api/http/api/mid"
	"github.com/ardanlabs/service/app/api/auth"
	"github.com/ardanlabs/service/app/api/authclient"
	"github.com/ardanlabs/service/app/domain/productapp"
	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log        *logger.Logger
	UserBus    *userbus.Business
	ProductBus *productbus.Business
	AuthClient *authclient.Client
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.Log, cfg.AuthClient)
	ruleAny := mid.Authorize(cfg.Log, cfg.AuthClient, auth.RuleAny)
	ruleUserOnly := mid.Authorize(cfg.Log, cfg.AuthClient, auth.RuleUserOnly)
	ruleAuthorizeProduct := mid.AuthorizeProduct(cfg.Log, cfg.AuthClient, cfg.ProductBus)

	api := newAPI(productapp.NewApp(cfg.ProductBus))
	app.Handle(http.MethodGet, version, "/products", api.query, authen, ruleAny)
	app.Handle(http.MethodGet, version, "/products/{product_id}", api.queryByID, authen, ruleAuthorizeProduct)
	app.Handle(http.MethodPost, version, "/products", api.create, authen, ruleUserOnly)
	app.Handle(http.MethodPut, version, "/products/{product_id}", api.update, authen, ruleAuthorizeProduct)
	app.Handle(http.MethodDelete, version, "/products/{product_id}", api.delete, authen, ruleAuthorizeProduct)
}
