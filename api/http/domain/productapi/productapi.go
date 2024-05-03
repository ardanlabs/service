// Package productapi maintains the web based api for product access.
package productapi

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/app/domain/productapp"
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

func (api *api) create(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
	var app productapp.NewProduct
	if err := web.Decode(r, &app); err != nil {
		return nil, errs.New(errs.FailedPrecondition, err)
	}

	prd, err := api.productApp.Create(ctx, app)
	if err != nil {
		return nil, err
	}

	return prd, nil
}

func (api *api) update(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
	var app productapp.UpdateProduct
	if err := web.Decode(r, &app); err != nil {
		return nil, errs.New(errs.FailedPrecondition, err)
	}

	prd, err := api.productApp.Update(ctx, app)
	if err != nil {
		return nil, err
	}

	return prd, nil
}

func (api *api) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
	if err := api.productApp.Delete(ctx); err != nil {
		return nil, err
	}

	return nil, nil
}

func (api *api) query(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
	qp, err := parseQueryParams(r)
	if err != nil {
		return nil, errs.New(errs.FailedPrecondition, err)
	}

	prd, err := api.productApp.Query(ctx, qp)
	if err != nil {
		return nil, err
	}

	return prd, nil
}

func (api *api) queryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
	prd, err := api.productApp.QueryByID(ctx)
	if err != nil {
		return nil, err
	}

	return prd, nil
}
