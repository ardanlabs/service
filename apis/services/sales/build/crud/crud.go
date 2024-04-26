// Package crud binds the crud domain set of routes into the specified app.
package crud

import (
	"github.com/ardanlabs/service/apis/services/sales/domain/checkapi"
	"github.com/ardanlabs/service/apis/services/sales/domain/homeapi"
	"github.com/ardanlabs/service/apis/services/sales/domain/productapi"
	"github.com/ardanlabs/service/apis/services/sales/domain/tranapi"
	"github.com/ardanlabs/service/apis/services/sales/domain/userapi"
	"github.com/ardanlabs/service/apis/services/sales/mux"
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

	homeapi.Routes(app, homeapi.Config{
		UserBus:    cfg.BusDomain.User,
		HomeBus:    cfg.BusDomain.Home,
		AuthClient: cfg.AuthClient,
	})

	productapi.Routes(app, productapi.Config{
		UserBus:    cfg.BusDomain.User,
		ProductBus: cfg.BusDomain.Product,
		AuthClient: cfg.AuthClient,
	})

	tranapi.Routes(app, tranapi.Config{
		UserBus:    cfg.BusDomain.User,
		ProductBus: cfg.BusDomain.Product,
		Log:        cfg.Log,
		AuthClient: cfg.AuthClient,
		DB:         cfg.DB,
	})

	userapi.Routes(app, userapi.Config{
		UserBus:    cfg.BusDomain.User,
		AuthClient: cfg.AuthClient,
	})
}
