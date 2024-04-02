// Package all binds all the routes into the specified app.
package all

import (
	"github.com/ardanlabs/service/apis/services/sales-api/http/routes/crud/homegrp"
	"github.com/ardanlabs/service/apis/services/sales-api/http/routes/crud/productgrp"
	"github.com/ardanlabs/service/apis/services/sales-api/http/routes/crud/trangrp"
	"github.com/ardanlabs/service/apis/services/sales-api/http/routes/crud/usergrp"
	"github.com/ardanlabs/service/apis/services/sales-api/http/routes/sys/checkgrp"
	"github.com/ardanlabs/service/apis/services/sales-api/http/routes/views/vproductgrp"
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
	checkgrp.Routes(app, checkgrp.Config{
		Build: cfg.Build,
		Log:   cfg.Log,
		DB:    cfg.DB,
	})

	homegrp.Routes(app, homegrp.Config{
		HomeCore: cfg.BusCrud.Home,
		Auth:     cfg.Auth,
	})

	productgrp.Routes(app, productgrp.Config{
		ProductCore: cfg.BusCrud.Product,
		Auth:        cfg.Auth,
	})

	trangrp.Routes(app, trangrp.Config{
		UserCore:    cfg.BusCrud.User,
		ProductCore: cfg.BusCrud.Product,
		Log:         cfg.Log,
		Auth:        cfg.Auth,
		DB:          cfg.DB,
	})

	usergrp.Routes(app, usergrp.Config{
		UserCore: cfg.BusCrud.User,
		Auth:     cfg.Auth,
	})

	vproductgrp.Routes(app, vproductgrp.Config{
		VProductCore: cfg.BusView.Product,
		Auth:         cfg.Auth,
	})
}
