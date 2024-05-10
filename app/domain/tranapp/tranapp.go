// Package tranapp maintains the app layer api for the tran domain.
package tranapp

import (
	"context"
	"errors"

	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/app/sdk/mid"
	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/ardanlabs/service/business/domain/userbus"
)

// App manages the set of app layer api functions for the tran domain.
type App struct {
	userBus    *userbus.Business
	productBus *productbus.Business
}

// NewApp constructs a tran app API for use.
func NewApp(userBus *userbus.Business, productBus *productbus.Business) *App {
	return &App{
		userBus:    userBus,
		productBus: productBus,
	}
}

// newWithTx constructs a new Handlers value with the domain apis
// using a store transaction that was created via middleware.
func (a *App) newWithTx(ctx context.Context) (*App, error) {
	tx, err := mid.GetTran(ctx)
	if err != nil {
		return nil, err
	}

	userBus, err := a.userBus.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	productBus, err := a.productBus.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	app := App{
		userBus:    userBus,
		productBus: productBus,
	}

	return &app, nil
}

// Create adds a new user and product at the same time under a single transaction.
func (a *App) Create(ctx context.Context, nt NewTran) (Product, error) {
	a, err := a.newWithTx(ctx)
	if err != nil {
		return Product{}, errs.New(errs.Internal, err)
	}

	np, err := toBusNewProduct(nt.Product)
	if err != nil {
		return Product{}, errs.New(errs.FailedPrecondition, err)
	}

	nu, err := toBusNewUser(nt.User)
	if err != nil {
		return Product{}, errs.New(errs.FailedPrecondition, err)
	}

	usr, err := a.userBus.Create(ctx, nu)
	if err != nil {
		if errors.Is(err, userbus.ErrUniqueEmail) {
			return Product{}, errs.New(errs.Aborted, userbus.ErrUniqueEmail)
		}
		return Product{}, errs.Newf(errs.Internal, "create: usr[%+v]: %s", usr, err)
	}

	np.UserID = usr.ID

	prd, err := a.productBus.Create(ctx, np)
	if err != nil {
		return Product{}, errs.Newf(errs.Internal, "create: prd[%+v]: %s", prd, err)
	}

	return toAppProduct(prd), nil
}
