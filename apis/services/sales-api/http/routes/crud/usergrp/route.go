package usergrp

import (
	"net/http"

	"github.com/ardanlabs/service/app/core/crud/userapp"
	"github.com/ardanlabs/service/business/api/auth"
	midhttp "github.com/ardanlabs/service/business/api/mid/http"
	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	UserCore *user.Core
	Auth     *auth.Auth
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := midhttp.Authenticate(cfg.Auth)
	ruleAdmin := midhttp.Authorize(cfg.Auth, auth.RuleAdminOnly)
	ruleAuthorizeUser := midhttp.AuthorizeUser(cfg.Auth, cfg.UserCore, auth.RuleAdminOrSubject)
	ruleAuthorizeAdmin := midhttp.AuthorizeUser(cfg.Auth, cfg.UserCore, auth.RuleAdminOnly)

	hdl := new(userapp.New(cfg.UserCore, cfg.Auth))
	app.Handle(http.MethodGet, version, "/users/token/{kid}", hdl.token)
	app.Handle(http.MethodGet, version, "/users", hdl.query, authen, ruleAdmin)
	app.Handle(http.MethodGet, version, "/users/{user_id}", hdl.queryByID, authen, ruleAuthorizeUser)
	app.Handle(http.MethodPost, version, "/users", hdl.create, authen, ruleAdmin)
	app.Handle(http.MethodPut, version, "/users/role/{user_id}", hdl.updateRole, authen, ruleAuthorizeAdmin)
	app.Handle(http.MethodPut, version, "/users/{user_id}", hdl.update, authen, ruleAuthorizeUser)
	app.Handle(http.MethodDelete, version, "/users/{user_id}", hdl.delete, authen, ruleAuthorizeUser)
}
