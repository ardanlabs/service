// Package productgrp maintains the group of handlers for product access.
package productgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/data/order"
	"github.com/ardanlabs/service/business/sys/validate"
	"github.com/ardanlabs/service/business/web/auth"
	v1Web "github.com/ardanlabs/service/business/web/v1"
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
func (h Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var np product.NewProduct
	if err := web.Decode(r, &np); err != nil {
		return err
	}

	prd, err := h.Product.Create(ctx, np)
	if err != nil {
		return fmt.Errorf("create: np[%+v]: %w", np, err)
	}

	return web.Respond(ctx, w, prd, http.StatusCreated)
}

// Update updates a product in the system.
func (h Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var up product.UpdateProduct
	if err := web.Decode(r, &up); err != nil {
		return err
	}

	id, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return validate.NewFieldsError("id", err)
	}

	prd, err := h.Product.QueryByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, product.ErrNotFound):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("querybyid: id[%s]: %w", id, err)
		}
	}

	prd, err = h.Product.Update(ctx, prd, up)
	if err != nil {
		return fmt.Errorf("update: id[%s] upd[%+v]: %w", id, up, err)
	}

	return web.Respond(ctx, w, prd, http.StatusOK)
}

// Delete removes a product from the system.
func (h Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return validate.NewFieldsError("id", err)
	}

	prd, err := h.Product.QueryByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, product.ErrNotFound):

			// Don't send StatusNotFound here since the call to Delete
			// below won't if this product is not found. We only know
			// this because we are doing the Query for the UserID.
			return v1Web.NewRequestError(err, http.StatusNoContent)
		default:
			return fmt.Errorf("querybyid: id[%s]: %w", id, err)
		}
	}

	if err := h.Product.Delete(ctx, prd); err != nil {
		return fmt.Errorf("delete: id[%s]: %w", id, err)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Query returns a list of products with paging.
func (h Handlers) Query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page := web.Param(r, "page")
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		return validate.NewFieldsError("page", err)
	}
	rows := web.Param(r, "rows")
	rowsPerPage, err := strconv.Atoi(rows)
	if err != nil {
		return validate.NewFieldsError("rows", err)
	}

	filter, err := parseFilter(r)
	if err != nil {
		return err
	}

	orderBy, err := order.Parse(r, h.Product.OrderingFields(), h.Product.DefaultOrderBy)
	if err != nil {
		return err
	}

	products, err := h.Product.Query(ctx, filter, orderBy, pageNumber, rowsPerPage)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	return web.Respond(ctx, w, products, http.StatusOK)
}

// QueryByID returns a product by its ID.
func (h Handlers) QueryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return validate.NewFieldsError("id", err)
	}

	prod, err := h.Product.QueryByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, product.ErrNotFound):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("querybyid: id[%s]: %w", id, err)
		}
	}

	return web.Respond(ctx, w, prod, http.StatusOK)
}
