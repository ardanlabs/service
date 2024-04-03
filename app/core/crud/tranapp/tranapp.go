// Package tranapp maintains the app layer api for the tran domain.
package tranapp

import (
	"context"
	"errors"

	"github.com/ardanlabs/service/business/api/errs"
	"github.com/ardanlabs/service/business/core/crud/productbus"
	"github.com/ardanlabs/service/business/core/crud/userbus"
)

// Core manages the set of app layer api functions for the tran domain.
type Core struct {
	userBus    *userbus.Core
	productBus *productbus.Core
}

// NewCore constructs a tran core API for use.
func NewCore(userBus *userbus.Core, productBus *productbus.Core) *Core {
	return &Core{
		userBus:    userBus,
		productBus: productBus,
	}
}

// Create adds a new user and product at the same time under a single transaction.
func (c *Core) Create(ctx context.Context, app NewTran) (Product, error) {
	api, err := c.executeUnderTransaction(ctx)
	if err != nil {
		return Product{}, errs.New(errs.Internal, err)
	}

	np, err := toBusNewProduct(app.Product)
	if err != nil {
		return Product{}, errs.New(errs.FailedPrecondition, err)
	}

	nu, err := toBusNewUser(app.User)
	if err != nil {
		return Product{}, errs.New(errs.FailedPrecondition, err)
	}

	usr, err := api.userBus.Create(ctx, nu)
	if err != nil {
		if errors.Is(err, userbus.ErrUniqueEmail) {
			return Product{}, errs.New(errs.Aborted, userbus.ErrUniqueEmail)
		}
		return Product{}, errs.Newf(errs.Internal, "create: usr[%+v]: %s", usr, err)
	}

	np.UserID = usr.ID

	prd, err := api.productBus.Create(ctx, np)
	if err != nil {
		return Product{}, errs.Newf(errs.Internal, "create: prd[%+v]: %s", prd, err)
	}

	return toAppProduct(prd), nil
}
