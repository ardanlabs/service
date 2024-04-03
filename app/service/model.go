package service

import (
	"github.com/ardanlabs/service/business/core/crud/delegate"
	"github.com/ardanlabs/service/business/core/crud/homebus"
	"github.com/ardanlabs/service/business/core/crud/productbus"
	"github.com/ardanlabs/service/business/core/crud/userbus"
	"github.com/ardanlabs/service/business/core/views/vproductbus"
)

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
