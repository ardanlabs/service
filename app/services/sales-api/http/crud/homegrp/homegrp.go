// Package homegrp maintains the group of handlers for home access.
package homegrp

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/services/sales-api/apis/crud/homeapi"
	"github.com/ardanlabs/service/business/api/errs"
	"github.com/ardanlabs/service/foundation/web"
)

type handlers struct {
	home *homeapi.API
}

func new(home *homeapi.API) *handlers {
	return &handlers{
		home: home,
	}
}

func (h *handlers) create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app homeapi.AppNewHome
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	hme, err := h.home.Create(ctx, app)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, hme, http.StatusCreated)
}

func (h *handlers) update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app homeapi.AppUpdateHome
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	hme, err := h.home.Update(ctx, app)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, hme, http.StatusOK)
}

func (h *handlers) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if err := h.home.Delete(ctx); err != nil {
		return err
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *handlers) query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	qp, err := parseQueryParams(r)
	if err != nil {
		return err
	}

	hme, err := h.home.Query(ctx, qp)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, hme, http.StatusOK)
}

func (h *handlers) queryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	hme, err := h.home.QueryByID(ctx)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, hme, http.StatusOK)
}
