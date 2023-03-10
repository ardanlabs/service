// Package productgrp maintains the group of handlers for product access.
package productgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/sys/validate"
	"github.com/ardanlabs/service/business/web/auth"
	v1Web "github.com/ardanlabs/service/business/web/v1"
	"github.com/ardanlabs/service/business/web/v1/paging"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/google/uuid"
)

// Set of error variables for handling product group errors.
var (
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// Handlers manages the set of product endpoints.
type Handlers struct {
	Product *product.Core
	Auth    *auth.Auth
}

// Create adds a new product to the system.
func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app AppNewProduct
	if err := web.Decode(r, &app); err != nil {
		return err
	}

	np, err := toCoreNewProduct(app)
	if err != nil {
		return v1Web.NewRequestError(err, http.StatusBadRequest)
	}

	prd, err := h.Product.Create(ctx, np)
	if err != nil {
		return fmt.Errorf("create: app[%+v]: %w", app, err)
	}

	return web.Respond(ctx, w, prd, http.StatusCreated)
}

// Update updates a product in the system.
func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app AppUpdateProduct
	if err := web.Decode(r, &app); err != nil {
		return err
	}

	productID, err := uuid.Parse(web.Param(r, "product_id"))
	if err != nil {
		return validate.NewFieldsError("product_id", err)
	}

	prd, err := h.Product.QueryByID(ctx, productID)
	if err != nil {
		switch {
		case errors.Is(err, product.ErrNotFound):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("querybyid: productID[%s]: %w", productID, err)
		}
	}

	prd, err = h.Product.Update(ctx, prd, toCoreUpdateProduct(app))
	if err != nil {
		return fmt.Errorf("update: productID[%s] app[%+v]: %w", productID, app, err)
	}

	return web.Respond(ctx, w, prd, http.StatusOK)
}

// Delete removes a product from the system.
func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	productID, err := uuid.Parse(web.Param(r, "product_id"))
	if err != nil {
		return validate.NewFieldsError("product_id", err)
	}

	prd, err := h.Product.QueryByID(ctx, productID)
	if err != nil {
		switch {
		case errors.Is(err, product.ErrNotFound):

			// Don't send StatusNotFound here since the call to Delete
			// below won't if this product is not found. We only know
			// this because we are doing the Query for the UserID.
			return v1Web.NewRequestError(err, http.StatusNoContent)
		default:
			return fmt.Errorf("querybyid: productID[%s]: %w", productID, err)
		}
	}

	if err := h.Product.Delete(ctx, prd); err != nil {
		return fmt.Errorf("delete: productID[%s]: %w", productID, err)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Query returns a list of products with paging.
func (h *Handlers) Query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page, err := paging.ParseRequest(r)
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

	prds, err := h.Product.Query(ctx, filter, orderBy, page.Number, page.RowsPerPage)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	items := make([]AppProduct, len(prds))
	for i, prd := range prds {
		items[i] = toAppProduct(prd)
	}

	total, err := h.Product.Count(ctx, filter)
	if err != nil {
		return fmt.Errorf("count: %w", err)
	}

	return web.Respond(ctx, w, paging.NewResponse(items, total, page.Number, page.RowsPerPage), http.StatusOK)
}

// QueryByID returns a product by its ID.
func (h *Handlers) QueryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	productID, err := uuid.Parse(web.Param(r, "product_id"))
	if err != nil {
		return validate.NewFieldsError("product_id", err)
	}

	prd, err := h.Product.QueryByID(ctx, productID)
	if err != nil {
		switch {
		case errors.Is(err, product.ErrNotFound):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("querybyid: productID[%s]: %w", productID, err)
		}
	}

	return web.Respond(ctx, w, toAppProduct(prd), http.StatusOK)
}
