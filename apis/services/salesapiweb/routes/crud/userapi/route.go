package userapi

import (
	"net/http"

	midhttp "github.com/ardanlabs/service/app/api/mid/http"
	"github.com/ardanlabs/service/app/core/crud/userapp"
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/core/crud/userbus"
	"github.com/ardanlabs/service/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	UserBus *userbus.Core
	Auth    *auth.Auth
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := midhttp.Authenticate(cfg.Auth)
	ruleAdmin := midhttp.Authorize(cfg.Auth, auth.RuleAdminOnly)
	ruleAuthorizeUser := midhttp.AuthorizeUser(cfg.Auth, cfg.UserBus, auth.RuleAdminOrSubject)
	ruleAuthorizeAdmin := midhttp.AuthorizeUser(cfg.Auth, cfg.UserBus, auth.RuleAdminOnly)

	api := newAPI(userapp.NewCore(cfg.UserBus, cfg.Auth))
	app.Handle(http.MethodGet, version, "/users/token/{kid}", api.token)
	app.Handle(http.MethodGet, version, "/users", api.query, authen, ruleAdmin)
	app.Handle(http.MethodGet, version, "/users/{user_id}", api.queryByID, authen, ruleAuthorizeUser)
	app.Handle(http.MethodPost, version, "/users", api.create, authen, ruleAdmin)
	app.Handle(http.MethodPut, version, "/users/role/{user_id}", api.updateRole, authen, ruleAuthorizeAdmin)
	app.Handle(http.MethodPut, version, "/users/{user_id}", api.update, authen, ruleAuthorizeUser)
	app.Handle(http.MethodDelete, version, "/users/{user_id}", api.delete, authen, ruleAuthorizeUser)
}
