// Package productapi maintains the web based api for product access.
package productapi

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/domain/productapp"
	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/foundation/web"
)

type api struct {
	productApp *productapp.App
}

func newAPI(productApp *productapp.App) *api {
	return &api{
		productApp: productApp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app productapp.NewProduct
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	prd, err := api.productApp.Create(ctx, app)
	if err != nil {
		return err.(*errs.Error)
	}

	return prd
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app productapp.UpdateProduct
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	prd, err := api.productApp.Update(ctx, app)
	if err != nil {
		return err.(*errs.Error)
	}

	return prd
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	if err := api.productApp.Delete(ctx); err != nil {
		return err.(*errs.Error)
	}

	return nil
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp := parseQueryParams(r)

	prd, err := api.productApp.Query(ctx, qp)
	if err != nil {
		return err.(*errs.Error)
	}

	return prd
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	prd, err := api.productApp.QueryByID(ctx)
	if err != nil {
		return err.(*errs.Error)
	}

	return prd
}
