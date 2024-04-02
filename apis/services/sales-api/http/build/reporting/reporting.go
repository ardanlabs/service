// Package reporting binds the reporting domain set of routes into the specified app.
package reporting

import (
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

	vproductgrp.Routes(app, vproductgrp.Config{
		VProductCore: cfg.BusView.Product,
		Auth:         cfg.Auth,
	})
}
