package tranapi

import (
	"net/http"

	midhttp "github.com/ardanlabs/service/app/api/mid/http"
	"github.com/ardanlabs/service/app/core/crud/tranapp"
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/core/crud/productbus"
	"github.com/ardanlabs/service/business/core/crud/userbus"
	"github.com/ardanlabs/service/business/data/sqldb"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/jmoiron/sqlx"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	UserBus    *userbus.Core
	ProductBus *productbus.Core
	Log        *logger.Logger
	DB         *sqlx.DB
	Auth       *auth.Auth
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := midhttp.Authenticate(cfg.UserBus, cfg.Auth)
	tran := midhttp.ExecuteInTransaction(cfg.Log, sqldb.NewBeginner(cfg.DB))

	api := newAPI(tranapp.NewCore(cfg.UserBus, cfg.ProductBus))
	app.Handle(http.MethodPost, version, "/tranexample", api.create, authen, tran)
}
