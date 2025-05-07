package dbtest

import (
	"time"

	"github.com/ardanlabs/service/business/domain/auditbus"
	"github.com/ardanlabs/service/business/domain/auditbus/stores/auditdb"
	"github.com/ardanlabs/service/business/domain/homebus"
	"github.com/ardanlabs/service/business/domain/homebus/stores/homedb"
	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/ardanlabs/service/business/domain/productbus/stores/productdb"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/business/domain/userbus/extensions/useraudit"
	"github.com/ardanlabs/service/business/domain/userbus/extensions/userotel"
	"github.com/ardanlabs/service/business/domain/userbus/stores/usercache"
	"github.com/ardanlabs/service/business/domain/userbus/stores/userdb"
	"github.com/ardanlabs/service/business/domain/vproductbus"
	"github.com/ardanlabs/service/business/domain/vproductbus/stores/vproductdb"
	"github.com/ardanlabs/service/business/sdk/delegate"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/jmoiron/sqlx"
)

// BusDomain represents all the business domain apis needed for testing.
type BusDomain struct {
	Delegate *delegate.Delegate
	Audit    *auditbus.Business
	Home     *homebus.Business
	Product  *productbus.Business
	User     userbus.ExtBusiness
	VProduct *vproductbus.Business
}

func newBusDomains(log *logger.Logger, db *sqlx.DB) BusDomain {
	userOtelExt := userotel.NewExtension()
	userAuditExt := useraudit.NewExtension(auditbus.NewBusiness(log, auditdb.NewStore(log, db)))
	userStorage := usercache.NewStore(log, userdb.NewStore(log, db), time.Hour)

	delegate := delegate.New(log)
	auditBus := auditbus.NewBusiness(log, auditdb.NewStore(log, db))
	userBus := userbus.NewBusiness(log, delegate, userStorage, userOtelExt, userAuditExt)
	productBus := productbus.NewBusiness(log, userBus, delegate, productdb.NewStore(log, db))
	homeBus := homebus.NewBusiness(log, userBus, delegate, homedb.NewStore(log, db))
	vproductBus := vproductbus.NewBusiness(vproductdb.NewStore(log, db))

	return BusDomain{
		Delegate: delegate,
		Audit:    auditBus,
		Home:     homeBus,
		Product:  productBus,
		User:     userBus,
		VProduct: vproductBus,
	}
}
