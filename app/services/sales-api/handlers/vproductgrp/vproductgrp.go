// Package vproductgrp maintains the group of handlers for detailed product data.
package vproductgrp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/business/core/views/vproduct"
	"github.com/ardanlabs/service/business/web/page"
	"github.com/ardanlabs/service/foundation/web"
)

type handlers struct {
	vProduct *vproduct.Core
}

func new(vProduct *vproduct.Core) *handlers {
	return &handlers{
		vProduct: vProduct,
	}
}

// Query returns a list of products with paging.
func (h *handlers) Query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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

	prds, err := h.vProduct.Query(ctx, filter, orderBy, pg.Number, pg.RowsPerPage)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	total, err := h.vProduct.Count(ctx, filter)
	if err != nil {
		return fmt.Errorf("count: %w", err)
	}

	return web.Respond(ctx, w, page.NewDocument(toAppProducts(prds), total, pg.Number, pg.RowsPerPage), http.StatusOK)
}
