// Package productapi maintains the web based api for product access.
package productapi

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/app/core/crud/productapp"
	"github.com/ardanlabs/service/foundation/web"
)

type api struct {
	product *productapp.Core
}

func newAPI(product *productapp.Core) *api {
	return &api{
		product: product,
	}
}

func (api *api) create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app productapp.NewProduct
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	prd, err := api.product.Create(ctx, app)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, prd, http.StatusCreated)
}

func (api *api) update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app productapp.UpdateProduct
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	prd, err := api.product.Update(ctx, app)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, prd, http.StatusOK)
}

func (api *api) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if err := api.product.Delete(ctx); err != nil {
		return err
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (api *api) query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	qp, err := parseQueryParams(r)
	if err != nil {
		return err
	}

	hme, err := api.product.Query(ctx, qp)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, hme, http.StatusOK)
}

func (api *api) queryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	hme, err := api.product.QueryByID(ctx)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, hme, http.StatusOK)
}
