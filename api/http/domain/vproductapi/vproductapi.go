// Package vproductapi maintains the web based api for product view access.
package vproductapi

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/domain/vproductapp"
	"github.com/ardanlabs/service/app/sdk/errs"
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

func (api *api) query(ctx context.Context, r *http.Request) (web.Encoder, error) {
	qp, err := parseQueryParams(r)
	if err != nil {
		return nil, errs.New(errs.FailedPrecondition, err)
	}

	prd, err := api.vproductApp.Query(ctx, qp)
	if err != nil {
		return nil, err
	}

	return prd, nil
}
