// Package productgrp maintains the group of handlers for product access.
package productgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/business/core/crud/product"
	"github.com/ardanlabs/service/business/web/errs"
	"github.com/ardanlabs/service/business/web/mid"
	"github.com/ardanlabs/service/business/web/page"
	"github.com/ardanlabs/service/foundation/web"
)

// Set of error variables for handling product group errors.
var (
	ErrInvalidID = errors.New("ID is not in its proper form")
)

type handlers struct {
	product *product.Core
}

func new(product *product.Core) *handlers {
	return &handlers{
		product: product,
	}
}

// create adds a new product to the system.
func (h *handlers) create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app AppNewProduct
	if err := web.Decode(r, &app); err != nil {
		return errs.NewTrusted(err, http.StatusBadRequest)
	}

	np, err := toCoreNewProduct(ctx, app)
	if err != nil {
		return errs.NewTrusted(err, http.StatusBadRequest)
	}

	prd, err := h.product.Create(ctx, np)
	if err != nil {
		return fmt.Errorf("create: app[%+v]: %w", app, err)
	}

	return web.Respond(ctx, w, toAppProduct(prd), http.StatusCreated)
}

// update updates a product in the system.
func (h *handlers) update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app AppUpdateProduct
	if err := web.Decode(r, &app); err != nil {
		return errs.NewTrusted(err, http.StatusBadRequest)
	}

	prd, err := mid.GetProduct(ctx)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}

	updPrd, err := h.product.Update(ctx, prd, toCoreUpdateProduct(app))
	if err != nil {
		return fmt.Errorf("update: productID[%s] app[%+v]: %w", prd.ID, app, err)
	}

	return web.Respond(ctx, w, toAppProduct(updPrd), http.StatusOK)
}

// delete removes a product from the system.
func (h *handlers) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	prd, err := mid.GetProduct(ctx)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	if err := h.product.Delete(ctx, prd); err != nil {
		return fmt.Errorf("delete: productID[%s]: %w", prd.ID, err)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// query returns a list of products with paging.
func (h *handlers) query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	pg, err := page.Parse(r)
	if err != nil {
		return err
	}

	filter, err := parseFilter(r)
	if err != nil {
		return err
	}

	orderBy, err := parseOrder(r)
	if err != nil {
		return err
	}

	prds, err := h.product.Query(ctx, filter, orderBy, pg.Number, pg.RowsPerPage)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	total, err := h.product.Count(ctx, filter)
	if err != nil {
		return fmt.Errorf("count: %w", err)
	}

	return web.Respond(ctx, w, page.NewDocument(toAppProducts(prds), total, pg.Number, pg.RowsPerPage), http.StatusOK)
}

// queryByID returns a product by its ID.
func (h *handlers) queryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	prd, err := mid.GetProduct(ctx)
	if err != nil {
		return fmt.Errorf("querybyid: %w", err)
	}

	return web.Respond(ctx, w, toAppProduct(prd), http.StatusOK)
}
