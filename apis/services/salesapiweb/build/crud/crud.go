// Package crud binds the crud domain set of routes into the specified app.
package crud

import (
	"github.com/ardanlabs/service/apis/services/salesapiweb/routes/crud/homegrp"
	"github.com/ardanlabs/service/apis/services/salesapiweb/routes/crud/productgrp"
	"github.com/ardanlabs/service/apis/services/salesapiweb/routes/crud/trangrp"
	"github.com/ardanlabs/service/apis/services/salesapiweb/routes/crud/usergrp"
	"github.com/ardanlabs/service/apis/services/salesapiweb/routes/sys/checkgrp"
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
}
