// Package all binds all the routes into the specified app.
package all

import (
	"github.com/ardanlabs/service/apis/services/auth/http/mux"
	"github.com/ardanlabs/service/apis/services/auth/http/route/authapi"
	"github.com/ardanlabs/service/apis/services/auth/http/route/checkapi"
	"github.com/ardanlabs/service/foundation/web"
)

// Routes constructs the add value which provides the implementation of
// of RouteAdder for specifying what routes to bind to this instance.
func Routes() add {
	return add{}
}

type add struct{}

// Add implements the RouterAdder interface.
func (add) Add(app *web.App, cfg mux.Config) {
	checkapi.Routes(app, checkapi.Config{
		Build: cfg.Build,
		Log:   cfg.Log,
		DB:    cfg.DB,
	})

	authapi.Routes(app, authapi.Config{
		UserBus: cfg.BusCrud.User,
		Auth:    cfg.Auth,
	})
}
