package homeapi

import (
	"net/http"

	"github.com/ardanlabs/service/apis/services/sales/mid"
	"github.com/ardanlabs/service/app/api/authsrv"
	"github.com/ardanlabs/service/app/core/crud/homeapp"
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/core/crud/homebus"
	"github.com/ardanlabs/service/business/core/crud/userbus"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log     *logger.Logger
	UserBus *userbus.Core
	HomeBus *homebus.Core
	AuthSrv *authsrv.AuthSrv
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.Log, cfg.AuthSrv)
	ruleAny := mid.Authorize(cfg.Log, cfg.AuthSrv, auth.RuleAny)
	ruleUserOnly := mid.Authorize(cfg.Log, cfg.AuthSrv, auth.RuleUserOnly)
	ruleAuthorizeHome := mid.AuthorizeHome(cfg.Log, cfg.AuthSrv, cfg.HomeBus)

	api := newAPI(homeapp.NewCore(cfg.HomeBus))
	app.Handle(http.MethodGet, version, "/homes", api.query, authen, ruleAny)
	app.Handle(http.MethodGet, version, "/homes/{home_id}", api.queryByID, authen, ruleAuthorizeHome)
	app.Handle(http.MethodPost, version, "/homes", api.create, authen, ruleUserOnly)
	app.Handle(http.MethodPut, version, "/homes/{home_id}", api.update, authen, ruleAuthorizeHome)
	app.Handle(http.MethodDelete, version, "/homes/{home_id}", api.delete, authen, ruleAuthorizeHome)
}
