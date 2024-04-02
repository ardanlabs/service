// Package vproductgrp maintains the group of handlers for detailed product data.
package vproductgrp

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/services/sales-api/apis/views/vproductapi"
	"github.com/ardanlabs/service/foundation/web"
)

type handlers struct {
	product *vproductapi.Handlers
}

func new(product *vproductapi.Handlers) *handlers {
	return &handlers{
		product: product,
	}
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
