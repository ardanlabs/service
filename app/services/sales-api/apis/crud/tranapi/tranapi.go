// Package tranapi maintains the group of handlers for transaction example.
package tranapi

import (
	"context"
	"errors"

	"github.com/ardanlabs/service/business/api/errs"
	"github.com/ardanlabs/service/business/core/crud/product"
	"github.com/ardanlabs/service/business/core/crud/user"
)

// API manages the set of handler functions for this domain.
type API struct {
	user    *user.Core
	product *product.Core
}

// New constructs a Handlers for use.
func New(user *user.Core, product *product.Core) *API {
	return &API{
		user:    user,
		product: product,
	}
}

// Create adds a new user and product at the same time under a single transaction.
func (api *API) Create(ctx context.Context, app AppNewTran) (AppProduct, error) {
	api, err := api.executeUnderTransaction(ctx)
	if err != nil {
		return AppProduct{}, errs.New(errs.Internal, err)
	}

	np, err := toCoreNewProduct(app.Product)
	if err != nil {
		return AppProduct{}, errs.New(errs.FailedPrecondition, err)
	}

	nu, err := toCoreNewUser(app.User)
	if err != nil {
		return AppProduct{}, errs.New(errs.FailedPrecondition, err)
	}

	usr, err := api.user.Create(ctx, nu)
	if err != nil {
		if errors.Is(err, user.ErrUniqueEmail) {
			return AppProduct{}, errs.New(errs.Aborted, user.ErrUniqueEmail)
		}
		return AppProduct{}, errs.Newf(errs.Internal, "create: usr[%+v]: %s", usr, err)
	}

	np.UserID = usr.ID

	prd, err := api.product.Create(ctx, np)
	if err != nil {
		return AppProduct{}, errs.Newf(errs.Internal, "create: prd[%+v]: %s", prd, err)
	}

	return toAppProduct(prd), nil
}
