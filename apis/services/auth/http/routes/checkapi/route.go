package checkapi

import (
	"net/http"

	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/jmoiron/sqlx"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Build string
	Log   *logger.Logger
	DB    *sqlx.DB
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	api := newAPI(cfg.Build, cfg.Log, cfg.DB)
	app.HandleNoMiddleware(http.MethodGet, version, "/readiness", api.readiness)
	app.HandleNoMiddleware(http.MethodGet, version, "/liveness", api.liveness)
}
