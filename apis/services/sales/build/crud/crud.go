// Package crud binds the crud domain set of routes into the specified app.
package crud

import (
	"github.com/ardanlabs/service/apis/services/sales/mux"
	"github.com/ardanlabs/service/apis/services/sales/routes/crud/homeapi"
	"github.com/ardanlabs/service/apis/services/sales/routes/crud/productapi"
	"github.com/ardanlabs/service/apis/services/sales/routes/crud/tranapi"
	"github.com/ardanlabs/service/apis/services/sales/routes/crud/userapi"
	"github.com/ardanlabs/service/apis/services/sales/routes/sys/checkapi"
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
		UserBus: cfg.BusCrud.User,
		HomeBus: cfg.BusCrud.Home,
		AuthSrv: cfg.AuthSrv,
	})

	productapi.Routes(app, productapi.Config{
		UserBus:    cfg.BusCrud.User,
		ProductBus: cfg.BusCrud.Product,
		AuthSrv:    cfg.AuthSrv,
	})

	tranapi.Routes(app, tranapi.Config{
		UserBus:    cfg.BusCrud.User,
		ProductBus: cfg.BusCrud.Product,
		Log:        cfg.Log,
		AuthSrv:    cfg.AuthSrv,
		DB:         cfg.DB,
	})

	userapi.Routes(app, userapi.Config{
		UserBus: cfg.BusCrud.User,
		AuthSrv: cfg.AuthSrv,
	})
}
