// Package crud binds the crud domain set of routes into the specified app.
package crud

import (
	"github.com/ardanlabs/service/apis/services/sales/routes/crud/homeapi"
	"github.com/ardanlabs/service/apis/services/sales/routes/crud/productapi"
	"github.com/ardanlabs/service/apis/services/sales/routes/crud/tranapi"
	"github.com/ardanlabs/service/apis/services/sales/routes/crud/userapi"
	"github.com/ardanlabs/service/apis/services/sales/routes/sys/checkapi"
	"github.com/ardanlabs/service/app/api/mux"
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
		HomeBus: cfg.BusCrud.Home,
		Auth:    cfg.Auth,
	})

	productapi.Routes(app, productapi.Config{
		ProductBus: cfg.BusCrud.Product,
		Auth:       cfg.Auth,
	})

	tranapi.Routes(app, tranapi.Config{
		UserBus:    cfg.BusCrud.User,
		ProductBus: cfg.BusCrud.Product,
		Log:        cfg.Log,
		Auth:       cfg.Auth,
		DB:         cfg.DB,
	})

	userapi.Routes(app, userapi.Config{
		UserBus: cfg.BusCrud.User,
		Auth:    cfg.Auth,
	})
}
