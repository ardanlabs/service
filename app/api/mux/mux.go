// Package mux provides support to bind domain level routes
// to the application mux.
package mux

import (
	"net/http"
	"os"

	midhttp "github.com/ardanlabs/service/app/api/mid/http"
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/core/crud/delegate"
	"github.com/ardanlabs/service/business/core/crud/homebus"
	"github.com/ardanlabs/service/business/core/crud/productbus"
	"github.com/ardanlabs/service/business/core/crud/userbus"
	"github.com/ardanlabs/service/business/core/views/vproductbus"
	"github.com/ardanlabs/service/foundation/authapi"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/trace"
)

// Options represent optional parameters.
type Options struct {
	corsOrigin []string
}

// WithCORS provides configuration options for CORS.
func WithCORS(origins []string) func(opts *Options) {
	return func(opts *Options) {
		opts.corsOrigin = origins
	}
}

// BusCrud represents the set of core business packages.
type BusCrud struct {
	Delegate *delegate.Delegate
	Home     *homebus.Core
	Product  *productbus.Core
	User     *userbus.Core
}

// BusView represents the set of view business packages.
type BusView struct {
	Product *vproductbus.Core
}

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Build    string
	Shutdown chan os.Signal
	Log      *logger.Logger
	Auth     *auth.Auth
	AuthAPI  *authapi.AuthAPI
	DB       *sqlx.DB
	Tracer   trace.Tracer
	BusCrud  BusCrud
	BusView  BusView
}

// RouteAdder defines behavior that sets the routes to bind for an instance
// of the service.
type RouteAdder interface {
	Add(app *web.App, cfg Config)
}

// WebAPI constructs a http.Handler with all application routes bound.
func WebAPI(cfg Config, routeAdder RouteAdder, options ...func(opts *Options)) http.Handler {
	var opts Options
	for _, option := range options {
		option(&opts)
	}

	app := web.NewApp(
		cfg.Shutdown,
		cfg.Tracer,
		midhttp.Logger(cfg.Log),
		midhttp.Errors(cfg.Log),
		midhttp.Metrics(),
		midhttp.Panics(),
	)

	if len(opts.corsOrigin) > 0 {
		app.EnableCORS(midhttp.Cors(opts.corsOrigin))
	}

	routeAdder.Add(app, cfg)

	return app
}
