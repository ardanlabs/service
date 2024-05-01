package authapi

import (
	"net/http"

	"github.com/ardanlabs/service/api/http/api/mid"
	"github.com/ardanlabs/service/app/api/auth"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	UserBus *userbus.Business
	Auth    *auth.Auth
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	bearer := mid.Bearer(cfg.Auth)
	basic := mid.Basic(cfg.UserBus, cfg.Auth)

	api := newAPI(cfg.Auth)
	app.Handle(http.MethodGet, version, "/auth/token/{kid}", api.token, basic)
	app.Handle(http.MethodGet, version, "/auth/authenticate", api.authenticate, bearer)
	app.Handle(http.MethodPost, version, "/auth/authorize", api.authorize)
}
