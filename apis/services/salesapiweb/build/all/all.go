// Package all binds all the routes into the specified app.
package all

import (
	"github.com/ardanlabs/service/apis/services/salesapiweb/routes/crud/homeapi"
	"github.com/ardanlabs/service/apis/services/salesapiweb/routes/crud/productapi"
	"github.com/ardanlabs/service/apis/services/salesapiweb/routes/crud/tranapi"
	"github.com/ardanlabs/service/apis/services/salesapiweb/routes/crud/userapi"
	"github.com/ardanlabs/service/apis/services/salesapiweb/routes/sys/checkapi"
	"github.com/ardanlabs/service/apis/services/salesapiweb/routes/views/vproductapi"
	"github.com/ardanlabs/service/business/api/mux"
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
		HomeCore: cfg.BusCrud.Home,
		Auth:     cfg.Auth,
	})

	productapi.Routes(app, productapi.Config{
		ProductCore: cfg.BusCrud.Product,
		Auth:        cfg.Auth,
	})

	tranapi.Routes(app, tranapi.Config{
		UserCore:    cfg.BusCrud.User,
		ProductCore: cfg.BusCrud.Product,
		Log:         cfg.Log,
		Auth:        cfg.Auth,
		DB:          cfg.DB,
	})

	userapi.Routes(app, userapi.Config{
		UserCore: cfg.BusCrud.User,
		Auth:     cfg.Auth,
	})

	vproductapi.Routes(app, vproductapi.Config{
		VProductCore: cfg.BusView.Product,
		Auth:         cfg.Auth,
	})
}
