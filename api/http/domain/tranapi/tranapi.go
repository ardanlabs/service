// Package tranapi maintains the web based api for tran access.
package tranapi

import (
	"context"
	"net/http"

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

func (api *api) create(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
	var app tranapp.NewTran
	if err := web.Decode(r, &app); err != nil {
		return nil, errs.New(errs.FailedPrecondition, err)
	}

	prd, err := api.tranApp.Create(ctx, app)
	if err != nil {
		return nil, err
	}

	return prd, nil
}
