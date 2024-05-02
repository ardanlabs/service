// Package vproductapi maintains the web based api for product view access.
package vproductapi

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/api/http/api/response"
	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/app/domain/vproductapp"
	"github.com/ardanlabs/service/foundation/web"
)

type api struct {
	vproductApp *vproductapp.App
}

func newAPI(vproductApp *vproductapp.App) *api {
	return &api{
		vproductApp: vproductApp,
	}
}

func (api *api) query(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
	qp, err := parseQueryParams(r)
	if err != nil {
		return response.AppError(errs.Internal, err)
	}

	hme, err := api.vproductApp.Query(ctx, qp)
	if err != nil {
		return response.AppError(errs.Internal, err)
	}

	return response.Response(hme, http.StatusOK)
}
