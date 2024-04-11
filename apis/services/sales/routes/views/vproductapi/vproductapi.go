// Package vproductapi maintains the web based api for product view access.
package vproductapi

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/core/views/vproductapp"
	"github.com/ardanlabs/service/foundation/web"
)

type api struct {
	vproductApp *vproductapp.Core
}

func newAPI(vproductApp *vproductapp.Core) *api {
	return &api{
		vproductApp: vproductApp,
	}
}

func (api *api) query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	qp, err := parseQueryParams(r)
	if err != nil {
		return err
	}

	hme, err := api.vproductApp.Query(ctx, qp)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, hme, http.StatusOK)
}
