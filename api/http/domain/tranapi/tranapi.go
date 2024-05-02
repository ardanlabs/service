// Package tranapi maintains the web based api for tran access.
package tranapi

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/api/http/api/response"
	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/app/domain/tranapp"
	"github.com/ardanlabs/service/foundation/web"
)

type api struct {
	tranApp *tranapp.App
}

func newAPI(tranApp *tranapp.App) *api {
	return &api{
		tranApp: tranApp,
	}
}

func (api *api) create(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
	var app tranapp.NewTran
	if err := web.Decode(r, &app); err != nil {
		return response.AppError(errs.FailedPrecondition, err)
	}

	prd, err := api.tranApp.Create(ctx, app)
	if err != nil {
		return response.AppAPIError(err)
	}

	return response.Response(prd, http.StatusCreated)
}
