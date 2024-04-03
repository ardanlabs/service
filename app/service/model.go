package service

import (
	"github.com/ardanlabs/service/business/core/crud/delegate"
	"github.com/ardanlabs/service/business/core/crud/home"
	"github.com/ardanlabs/service/business/core/crud/product"
	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/core/views/vproduct"
)

// BusCrud represents the set of core business packages.
type BusCrud struct {
	Delegate *delegate.Delegate
	Home     *home.Core
	Product  *product.Core
	User     *user.Core
}

// BusView represents the set of view business packages.
type BusView struct {
	Product *vproduct.Core
}
