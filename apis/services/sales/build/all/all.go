// Package all binds all the routes into the specified app.
package all

import (
	"github.com/ardanlabs/service/apis/services/sales/domain/checkapi"
	"github.com/ardanlabs/service/apis/services/sales/domain/homeapi"
	"github.com/ardanlabs/service/apis/services/sales/domain/productapi"
	"github.com/ardanlabs/service/apis/services/sales/domain/tranapi"
	"github.com/ardanlabs/service/apis/services/sales/domain/userapi"
	"github.com/ardanlabs/service/apis/services/sales/domain/vproductapi"
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
		Log:        cfg.Log,
		UserBus:    cfg.BusDomain.User,
		HomeBus:    cfg.BusDomain.Home,
		AuthClient: cfg.AuthClient,
	})

	productapi.Routes(app, productapi.Config{
		Log:        cfg.Log,
		UserBus:    cfg.BusDomain.User,
		ProductBus: cfg.BusDomain.Product,
		AuthClient: cfg.AuthClient,
	})

	tranapi.Routes(app, tranapi.Config{
		Log:        cfg.Log,
		DB:         cfg.DB,
		UserBus:    cfg.BusDomain.User,
		ProductBus: cfg.BusDomain.Product,
		AuthClient: cfg.AuthClient,
	})

	userapi.Routes(app, userapi.Config{
		Log:        cfg.Log,
		UserBus:    cfg.BusDomain.User,
		AuthClient: cfg.AuthClient,
	})

	vproductapi.Routes(app, vproductapi.Config{
		Log:         cfg.Log,
		UserBus:     cfg.BusDomain.User,
		VProductBus: cfg.BusDomain.VProduct,
		AuthClient:  cfg.AuthClient,
	})
}
