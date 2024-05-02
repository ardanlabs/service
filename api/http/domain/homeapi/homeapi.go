// Package homeapi maintains the web based api for home access.
package homeapi

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/api/http/api/response"
	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/app/domain/homeapp"
	"github.com/ardanlabs/service/foundation/web"
)

type api struct {
	homeApp *homeapp.App
}

func newAPI(homeApp *homeapp.App) *api {
	return &api{
		homeApp: homeApp,
	}
}

func (api *api) create(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
	var app homeapp.NewHome
	if err := web.Decode(r, &app); err != nil {
		return response.AppError(errs.FailedPrecondition, err)
	}

	hme, err := api.homeApp.Create(ctx, app)
	if err != nil {
		return response.AppAPIError(err)
	}

	return response.Response(hme, http.StatusCreated)
}

func (api *api) update(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
	var app homeapp.UpdateHome
	if err := web.Decode(r, &app); err != nil {
		return response.AppError(errs.FailedPrecondition, err)
	}

	hme, err := api.homeApp.Update(ctx, app)
	if err != nil {
		return response.AppAPIError(err)
	}

	return response.Response(hme, http.StatusOK)
}

func (api *api) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
	if err := api.homeApp.Delete(ctx); err != nil {
		return response.AppAPIError(err)
	}

	return response.Response(nil, http.StatusNoContent)
}

func (api *api) query(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
	qp, err := parseQueryParams(r)
	if err != nil {
		return response.AppError(errs.FailedPrecondition, err)
	}

	hme, err := api.homeApp.Query(ctx, qp)
	if err != nil {
		return response.AppAPIError(err)
	}

	return response.Response(hme, http.StatusOK)
}

func (api *api) queryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
	hme, err := api.homeApp.QueryByID(ctx)
	if err != nil {
		return response.AppAPIError(err)
	}

	return response.Response(hme, http.StatusOK)
}
