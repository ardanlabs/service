package tranapi

import (
	"net/http"

	"github.com/ardanlabs/service/api/http/api/mid"
	"github.com/ardanlabs/service/app/api/authclient"
	"github.com/ardanlabs/service/app/domain/tranapp"
	"github.com/ardanlabs/service/business/api/sqldb"
	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/jmoiron/sqlx"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log        *logger.Logger
	DB         *sqlx.DB
	UserBus    *userbus.Business
	ProductBus *productbus.Business
	AuthClient *authclient.Client
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.Log, cfg.AuthClient)
	transaction := mid.BeginCommitRollback(cfg.Log, sqldb.NewBeginner(cfg.DB))

	api := newAPI(tranapp.NewApp(cfg.UserBus, cfg.ProductBus))
	app.Handle(http.MethodPost, version, "/tranexample", api.create, authen, transaction)
}
