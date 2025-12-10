package build

import (
	"github.com/ardanlabs/service/app/domain/authapp"
	"github.com/ardanlabs/service/app/domain/checkapp"
	"github.com/ardanlabs/service/app/sdk/mux"
	"github.com/ardanlabs/service/foundation/web"
)

// All binds all the routes for the auth service.
func All() all {
	return all{}
}

type all struct{}

// Add implements the RouterAdder interface.
func (all) Add(app *web.App, cfg mux.Config) {
	checkapp.Routes(app, checkapp.Config{
		Build: cfg.Build,
		Log:   cfg.Log,
		DB:    cfg.DB,
	})

	authapp.Routes(app, authapp.Config{
		UserBus: cfg.BusConfig.UserBus,
		Auth:    cfg.AuthConfig.Auth,
	})
}
