// Package all binds all the routes into the specified app.
package all

import (
	"github.com/ardanlabs/service/apis/api/mux"
	"github.com/ardanlabs/service/apis/domain/authapi"
	"github.com/ardanlabs/service/apis/domain/checkapi"
	"github.com/ardanlabs/service/business/api/delegate"
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

	checkapi.Routes(app, checkapi.Config{
		Build: cfg.Build,
		Log:   cfg.Log,
		DB:    cfg.DB,
	})

	authapi.Routes(app, authapi.Config{
		UserBus: userBus,
		Auth:    cfg.Auth,
	})
}
