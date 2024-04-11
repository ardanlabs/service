// Package all binds all the routes into the specified app.
package all

import (
	"github.com/ardanlabs/service/apis/services/sales/http/mux"
	"github.com/ardanlabs/service/apis/services/sales/http/routes/crud/homeapi"
	"github.com/ardanlabs/service/apis/services/sales/http/routes/crud/productapi"
	"github.com/ardanlabs/service/apis/services/sales/http/routes/crud/tranapi"
	"github.com/ardanlabs/service/apis/services/sales/http/routes/crud/userapi"
	"github.com/ardanlabs/service/apis/services/sales/http/routes/sys/checkapi"
	"github.com/ardanlabs/service/apis/services/sales/http/routes/views/vproductapi"
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
		Log:     cfg.Log,
		UserBus: cfg.BusCrud.User,
		HomeBus: cfg.BusCrud.Home,
		AuthSrv: cfg.AuthSrv,
	})

	productapi.Routes(app, productapi.Config{
		Log:        cfg.Log,
		UserBus:    cfg.BusCrud.User,
		ProductBus: cfg.BusCrud.Product,
		AuthSrv:    cfg.AuthSrv,
	})

	tranapi.Routes(app, tranapi.Config{
		Log:        cfg.Log,
		DB:         cfg.DB,
		UserBus:    cfg.BusCrud.User,
		ProductBus: cfg.BusCrud.Product,
		AuthSrv:    cfg.AuthSrv,
	})

	userapi.Routes(app, userapi.Config{
		Log:     cfg.Log,
		UserBus: cfg.BusCrud.User,
		AuthSrv: cfg.AuthSrv,
	})

	vproductapi.Routes(app, vproductapi.Config{
		Log:         cfg.Log,
		UserBus:     cfg.BusCrud.User,
		VProductBus: cfg.BusView.Product,
		AuthSrv:     cfg.AuthSrv,
	})
}
