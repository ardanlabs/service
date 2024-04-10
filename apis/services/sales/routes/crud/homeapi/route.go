package homeapi

import (
	"net/http"

	midhttp "github.com/ardanlabs/service/app/api/mid/http"
	"github.com/ardanlabs/service/app/core/crud/homeapp"
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/core/crud/homebus"
	"github.com/ardanlabs/service/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	HomeBus *homebus.Core
	Auth    *auth.Auth
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := midhttp.Authenticate(cfg.Auth)
	ruleAny := midhttp.Authorize(cfg.Auth, auth.RuleAny)
	ruleUserOnly := midhttp.Authorize(cfg.Auth, auth.RuleUserOnly)
	ruleAuthorizeHome := midhttp.AuthorizeHome(cfg.Auth, cfg.HomeBus)

	api := newAPI(homeapp.NewCore(cfg.HomeBus))
	app.Handle(http.MethodGet, version, "/homes", api.query, authen, ruleAny)
	app.Handle(http.MethodGet, version, "/homes/{home_id}", api.queryByID, authen, ruleAuthorizeHome)
	app.Handle(http.MethodPost, version, "/homes", api.create, authen, ruleUserOnly)
	app.Handle(http.MethodPut, version, "/homes/{home_id}", api.update, authen, ruleAuthorizeHome)
	app.Handle(http.MethodDelete, version, "/homes/{home_id}", api.delete, authen, ruleAuthorizeHome)
}
