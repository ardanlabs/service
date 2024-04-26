// Package crud binds the crud domain set of routes into the specified app.
package crud

import (
	"github.com/ardanlabs/service/api/http/api/mux"
	"github.com/ardanlabs/service/api/http/domain/checkapi"
	"github.com/ardanlabs/service/api/http/domain/homeapi"
	"github.com/ardanlabs/service/api/http/domain/productapi"
	"github.com/ardanlabs/service/api/http/domain/tranapi"
	"github.com/ardanlabs/service/api/http/domain/userapi"
	"github.com/ardanlabs/service/business/api/delegate"
	"github.com/ardanlabs/service/business/domain/homebus"
	"github.com/ardanlabs/service/business/domain/homebus/stores/homedb"
	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/ardanlabs/service/business/domain/productbus/stores/productdb"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/business/domain/userbus/stores/userdb"
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

	// Construct the business domain packages we need here so we are using the
	// sames instances for the different set of domain apis.
	delegate := delegate.New(cfg.Log)
	userBus := userbus.NewCore(cfg.Log, delegate, userdb.NewStore(cfg.Log, cfg.DB))
	productBus := productbus.NewCore(cfg.Log, userBus, delegate, productdb.NewStore(cfg.Log, cfg.DB))
	homeBus := homebus.NewCore(cfg.Log, userBus, delegate, homedb.NewStore(cfg.Log, cfg.DB))

	checkapi.Routes(app, checkapi.Config{
		Build: cfg.Build,
		Log:   cfg.Log,
		DB:    cfg.DB,
	})

	homeapi.Routes(app, homeapi.Config{
		UserBus:    userBus,
		HomeBus:    homeBus,
		AuthClient: cfg.AuthClient,
	})

	productapi.Routes(app, productapi.Config{
		UserBus:    userBus,
		ProductBus: productBus,
		AuthClient: cfg.AuthClient,
	})

	tranapi.Routes(app, tranapi.Config{
		UserBus:    userBus,
		ProductBus: productBus,
		Log:        cfg.Log,
		AuthClient: cfg.AuthClient,
		DB:         cfg.DB,
	})

	userapi.Routes(app, userapi.Config{
		UserBus:    userBus,
		AuthClient: cfg.AuthClient,
	})
}
