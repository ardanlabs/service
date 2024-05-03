// Package homeapi maintains the web based api for home access.
package homeapi

import (
	"context"
	"net/http"

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

func (api *api) create(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
	var app homeapp.NewHome
	if err := web.Decode(r, &app); err != nil {
		return nil, errs.New(errs.FailedPrecondition, err)
	}

	hme, err := api.homeApp.Create(ctx, app)
	if err != nil {
		return nil, err
	}

	return hme, nil
}

func (api *api) update(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
	var app homeapp.UpdateHome
	if err := web.Decode(r, &app); err != nil {
		return nil, errs.New(errs.FailedPrecondition, err)
	}

	hme, err := api.homeApp.Update(ctx, app)
	if err != nil {
		return nil, err
	}

	return hme, nil
}

func (api *api) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
	if err := api.homeApp.Delete(ctx); err != nil {
		return nil, err
	}

	return nil, nil
}

func (api *api) query(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
	qp, err := parseQueryParams(r)
	if err != nil {
		return nil, errs.New(errs.FailedPrecondition, err)
	}

	hme, err := api.homeApp.Query(ctx, qp)
	if err != nil {
		return nil, err
	}

	return hme, nil
}

func (api *api) queryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
	hme, err := api.homeApp.QueryByID(ctx)
	if err != nil {
		return nil, err
	}

	return hme, nil
}
