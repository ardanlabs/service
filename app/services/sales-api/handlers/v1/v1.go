// Package v1 manages the different versions of the API.
package v1

import (
	"net/http"
	"os"

	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/groups/checkgrp"
	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/groups/homegrp"
	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/groups/productgrp"
	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/groups/trangrp"
	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/groups/usergrp"
	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/groups/usersummarygrp"
	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/mid"
	"github.com/ardanlabs/service/business/web/auth"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/trace"
)

// Options represent optional parameters.
type Options struct {
	corsOrigin string
}

// WithCORS provides configuration options for CORS.
func WithCORS(origin string) func(opts *Options) {
	return func(opts *Options) {
		opts.corsOrigin = origin
	}
}

// APIMuxConfig contains all the mandatory systems required by handlers.
type APIMuxConfig struct {
	UsingWeaver bool
	Build       string
	Shutdown    chan os.Signal
	Log         *logger.Logger
	Auth        *auth.Auth
	DB          *sqlx.DB
	Tracer      trace.Tracer
}

// APIMux constructs a http.Handler with all application routes defined.
func APIMux(cfg APIMuxConfig, options ...func(opts *Options)) http.Handler {
	var opts Options
	for _, option := range options {
		option(&opts)
	}

	app := web.NewApp(
		cfg.Shutdown,
		cfg.Tracer,
		mid.Logger(cfg.Log),
		mid.Errors(cfg.Log),
		mid.Metrics(),
		mid.Panics(),
	)

	if opts.corsOrigin != "" {
		app.EnableCORS(mid.Cors(opts.corsOrigin))
	}

	addRoutes(app, cfg)

	return app
}

func addRoutes(app *web.App, cfg APIMuxConfig) {
	checkgrp.Routes(app, checkgrp.Config{
		UsingWeaver: cfg.UsingWeaver,
		Build:       cfg.Build,
		DB:          cfg.DB,
	})

	homegrp.Routes(app, homegrp.Config{
		Log:  cfg.Log,
		Auth: cfg.Auth,
		DB:   cfg.DB,
	})

	productgrp.Routes(app, productgrp.Config{
		Log:  cfg.Log,
		Auth: cfg.Auth,
		DB:   cfg.DB,
	})

	trangrp.Routes(app, trangrp.Config{
		Log:  cfg.Log,
		Auth: cfg.Auth,
		DB:   cfg.DB,
	})

	usergrp.Routes(app, usergrp.Config{
		Log:  cfg.Log,
		Auth: cfg.Auth,
		DB:   cfg.DB,
	})

	usersummarygrp.Routes(app, usersummarygrp.Config{
		Log:  cfg.Log,
		Auth: cfg.Auth,
		DB:   cfg.DB,
	})
}
