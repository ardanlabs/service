// Package productapi maintains the web based api for product access.
package productapi

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/api/http/api/response"
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

func (api *api) create(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
	var app productapp.NewProduct
	if err := web.Decode(r, &app); err != nil {
		return response.AppError(errs.FailedPrecondition, err)
	}

	prd, err := api.productApp.Create(ctx, app)
	if err != nil {
		return response.AppError(errs.Internal, err)
	}

	return response.Response(prd, http.StatusCreated)
}

func (api *api) update(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
	var app productapp.UpdateProduct
	if err := web.Decode(r, &app); err != nil {
		return response.AppError(errs.FailedPrecondition, err)
	}

	prd, err := api.productApp.Update(ctx, app)
	if err != nil {
		return response.AppError(errs.Internal, err)
	}

	return response.Response(prd, http.StatusOK)
}

func (api *api) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
	if err := api.productApp.Delete(ctx); err != nil {
		return response.AppError(errs.Internal, err)
	}

	return response.Response(nil, http.StatusNoContent)
}

func (api *api) query(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
	qp, err := parseQueryParams(r)
	if err != nil {
		return response.AppError(errs.Internal, err)
	}

	hme, err := api.productApp.Query(ctx, qp)
	if err != nil {
		return response.AppError(errs.Internal, err)
	}

	return response.Response(hme, http.StatusOK)
}

func (api *api) queryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
	hme, err := api.productApp.QueryByID(ctx)
	if err != nil {
		return response.AppError(errs.Internal, err)
	}

	return response.Response(hme, http.StatusOK)
}
