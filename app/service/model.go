package service

import (
	"github.com/ardanlabs/service/app/core/crud/homeapp"
	"github.com/ardanlabs/service/app/core/crud/productapp"
	"github.com/ardanlabs/service/app/core/crud/tranapp"
	"github.com/ardanlabs/service/app/core/crud/userapp"
	"github.com/ardanlabs/service/app/core/views/vproductapp"
	"github.com/ardanlabs/service/business/core/crud/delegate"
	"github.com/ardanlabs/service/business/core/crud/home"
	"github.com/ardanlabs/service/business/core/crud/product"
	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/core/views/vproduct"
)

// CrudApp represents the set of crud based app layer APIs.
type CrudApp struct {
	Home    *homeapp.Core
	Product *productapp.Core
	Tran    *tranapp.Core
	User    *userapp.Core
}

// ViewApp represents the set of view based app layer APIs.
type ViewApp struct {
	Product *vproductapp.Core
}

// App represents the set of all app layer APIs.
type App struct {
	Crud CrudApp
	View ViewApp
}

// CrudBus represents the set of crud based business layer APIs.
type CrudBus struct {
	Delegate *delegate.Delegate
	Home     *home.Core
	Product  *product.Core
	User     *user.Core
}

// ViewBus represents the set of view based business layer APIs.
type ViewBus struct {
	Product *vproduct.Core
}

// Business represents the set of all business layer APIs.
type Business struct {
	Crud CrudBus
	View ViewBus
}
