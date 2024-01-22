package homegrp

import (
	"net/http"

	"github.com/ardanlabs/service/business/core/event"
	"github.com/ardanlabs/service/business/core/home"
	"github.com/ardanlabs/service/business/core/home/stores/homedb"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/core/user/stores/usercache"
	"github.com/ardanlabs/service/business/core/user/stores/userdb"
	"github.com/ardanlabs/service/business/web/v1/auth"
	"github.com/ardanlabs/service/business/web/v1/mid"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/jmoiron/sqlx"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log  *logger.Logger
	Auth *auth.Auth
	DB   *sqlx.DB
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	envCore := event.NewCore(cfg.Log)
	usrCore := user.NewCore(cfg.Log, envCore, usercache.NewStore(cfg.Log, userdb.NewStore(cfg.Log, cfg.DB)))
	hmeCore := home.NewCore(cfg.Log, envCore, usrCore, homedb.NewStore(cfg.Log, cfg.DB))

	authen := mid.Authenticate(cfg.Auth)
	ruleAny := mid.Authorize(cfg.Auth, auth.RuleAny)
	ruleUserOnly := mid.Authorize(cfg.Auth, auth.RuleUserOnly)
	ruleAdminOrSubject := mid.AuthorizeHome(cfg.Auth, auth.RuleAdminOrSubject, hmeCore)

	hdl := new(hmeCore)
	app.Handle(http.MethodGet, version, "/homes", hdl.query, authen, ruleAny)
	app.Handle(http.MethodGet, version, "/homes/:home_id", hdl.queryByID, authen, ruleAdminOrSubject)
	app.Handle(http.MethodPost, version, "/homes", hdl.create, authen, ruleUserOnly)
	app.Handle(http.MethodPut, version, "/homes/:home_id", hdl.update, authen, ruleAdminOrSubject)
	app.Handle(http.MethodDelete, version, "/homes/:home_id", hdl.delete, authen, ruleAdminOrSubject)
}
