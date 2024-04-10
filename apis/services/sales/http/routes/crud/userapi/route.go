package userapi

import (
	"net/http"

	midhttp "github.com/ardanlabs/service/app/api/mid/http"
	"github.com/ardanlabs/service/app/core/crud/userapp"
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/core/crud/userbus"
	"github.com/ardanlabs/service/foundation/authapi"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log     *logger.Logger
	UserBus *userbus.Core
	AuthAPI *authapi.AuthAPI
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := midhttp.AuthenticateWeb(cfg.Log, cfg.AuthAPI)
	ruleAdmin := midhttp.Authorize(cfg.Log, cfg.AuthAPI, auth.RuleAdminOnly)
	ruleAuthorizeUser := midhttp.AuthorizeUser(cfg.Log, cfg.AuthAPI, cfg.UserBus, auth.RuleAdminOrSubject)
	ruleAuthorizeAdmin := midhttp.AuthorizeUser(cfg.Log, cfg.AuthAPI, cfg.UserBus, auth.RuleAdminOnly)

	api := newAPI(userapp.NewCore(cfg.UserBus))
	app.Handle(http.MethodGet, version, "/users", api.query, authen, ruleAdmin)
	app.Handle(http.MethodGet, version, "/users/{user_id}", api.queryByID, authen, ruleAuthorizeUser)
	app.Handle(http.MethodPost, version, "/users", api.create, authen, ruleAdmin)
	app.Handle(http.MethodPut, version, "/users/role/{user_id}", api.updateRole, authen, ruleAuthorizeAdmin)
	app.Handle(http.MethodPut, version, "/users/{user_id}", api.update, authen, ruleAuthorizeUser)
	app.Handle(http.MethodDelete, version, "/users/{user_id}", api.delete, authen, ruleAuthorizeUser)
}
