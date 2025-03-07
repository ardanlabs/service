// Package reporting binds the reporting domain set of routes into the specified app.
package reporting

import (
	"github.com/ardanlabs/service/app/domain/checkapp"
	"github.com/ardanlabs/service/app/domain/vproductapp"
	"github.com/ardanlabs/service/app/sdk/mux"
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
	checkapp.Routes(app, checkapp.Config{
		Build: cfg.Build,
		Log:   cfg.Log,
		DB:    cfg.DB,
	})

	vproductapp.Routes(app, vproductapp.Config{
		UserBus:     cfg.BusConfig.UserBus,
		VProductBus: cfg.BusConfig.VProductBus,
		AuthClient:  cfg.SalesConfig.AuthClient,
	})
}
