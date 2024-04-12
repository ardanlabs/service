package tranapi

import (
	"net/http"

	"github.com/ardanlabs/service/apis/services/sales/mid"
	"github.com/ardanlabs/service/app/api/authsrv"
	midapp "github.com/ardanlabs/service/app/api/mid/http"
	"github.com/ardanlabs/service/app/domain/tranapp"
	"github.com/ardanlabs/service/business/data/sqldb"
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
	UserBus    *userbus.Core
	ProductBus *productbus.Core
	AuthSrv    *authsrv.AuthSrv
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.Log, cfg.AuthSrv)
	tran := midapp.ExecuteInTransaction(cfg.Log, sqldb.NewBeginner(cfg.DB))

	api := newAPI(tranapp.NewCore(cfg.UserBus, cfg.ProductBus))
	app.Handle(http.MethodPost, version, "/tranexample", api.create, authen, tran)
}
