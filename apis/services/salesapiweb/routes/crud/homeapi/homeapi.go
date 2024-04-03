// Package homeapi maintains the web based api for home access.
package homeapi

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/app/core/crud/homeapp"
	"github.com/ardanlabs/service/foundation/web"
)

type api struct {
	home *homeapp.Core
}

func newAPI(home *homeapp.Core) *api {
	return &api{
		home: home,
	}
}

func (api *api) create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app homeapp.NewHome
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	hme, err := api.home.Create(ctx, app)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, hme, http.StatusCreated)
}

func (api *api) update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app homeapp.UpdateHome
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	hme, err := api.home.Update(ctx, app)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, hme, http.StatusOK)
}

func (api *api) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if err := api.home.Delete(ctx); err != nil {
		return err
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (api *api) query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	qp, err := parseQueryParams(r)
	if err != nil {
		return err
	}

	hme, err := api.home.Query(ctx, qp)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, hme, http.StatusOK)
}

func (api *api) queryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	hme, err := api.home.QueryByID(ctx)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, hme, http.StatusOK)
}
