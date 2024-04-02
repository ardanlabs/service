// Package productgrp maintains the group of handlers for product access.
package productgrp

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/core/crud/productapp"
	"github.com/ardanlabs/service/business/api/errs"
	"github.com/ardanlabs/service/foundation/web"
)

type handlers struct {
	product *productapp.Core
}

func new(product *productapp.Core) *handlers {
	return &handlers{
		product: product,
	}
}

func (h *handlers) create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app productapp.NewProduct
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	prd, err := h.product.Create(ctx, app)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, prd, http.StatusCreated)
}

func (h *handlers) update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app productapp.UpdateProduct
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	prd, err := h.product.Update(ctx, app)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, prd, http.StatusOK)
}

func (h *handlers) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if err := h.product.Delete(ctx); err != nil {
		return err
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *handlers) query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	qp, err := parseQueryParams(r)
	if err != nil {
		return err
	}

	hme, err := h.product.Query(ctx, qp)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, hme, http.StatusOK)
}

func (h *handlers) queryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	hme, err := h.product.QueryByID(ctx)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, hme, http.StatusOK)
}
