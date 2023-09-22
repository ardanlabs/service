// Package product binds the product domain set of routes into the specified app.
package product

import (
	v1 "github.com/ardanlabs/service/app/services/sales-api/v1"
	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/checkgrp"
	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/productgrp"
	"github.com/ardanlabs/service/foundation/web"
)

// Routes constructs the add value which provides the implementation of
// of RouteAdder for specifying what routes to bind to this instance.
func Routes() add {
	return add{}
}

type add struct{}

// Add implements the RouterAdder interface.
func (add) Add(app *web.App, cfg v1.APIMuxConfig) {
	checkgrp.Routes(app, checkgrp.Config{
		UsingWeaver: cfg.UsingWeaver,
		Build:       cfg.Build,
		DB:          cfg.DB,
	})

	productgrp.Routes(app, productgrp.Config{
		Log:  cfg.Log,
		Auth: cfg.Auth,
		DB:   cfg.DB,
	})
}
