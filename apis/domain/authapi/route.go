package authapi

import (
	"net/http"

	"github.com/ardanlabs/service/apis/api/mid"
	"github.com/ardanlabs/service/app/domain/userapp"
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/domain/userbus"
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

	authen := mid.BearerBasic(cfg.UserBus, cfg.Auth)

	api := newAPI(userapp.NewCoreWithAuth(cfg.UserBus, cfg.Auth), cfg.Auth)
	app.Handle(http.MethodGet, version, "/auth/token/{kid}", api.token, authen)
	app.Handle(http.MethodGet, version, "/auth/authenticate", api.authenticate, authen)
	app.Handle(http.MethodPost, version, "/auth/authorize", api.authorize)
}
