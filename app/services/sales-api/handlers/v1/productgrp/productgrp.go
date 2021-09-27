// Package productgrp maintains the group of handlers for product access.
package productgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/sys/auth"
	"github.com/ardanlabs/service/business/sys/validate"
	webv1 "github.com/ardanlabs/service/business/web/v1"
	"github.com/ardanlabs/service/foundation/web"
)

// Handlers manages the set of product enpoints.
type Handlers struct {
	Core product.Core
}

// Create adds a new product to the system.
func (h Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	var np product.NewProduct
	if err := web.Decode(r, &np); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	prod, err := h.Core.Create(ctx, np, v.Now)
	if err != nil {
		return fmt.Errorf("creating new product, np[%+v]: %w", np, err)
	}

	return web.Respond(ctx, w, prod, http.StatusCreated)
}

// Update updates a product in the system.
func (h Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return errors.New("claims missing from context")
	}

	var upd product.UpdateProduct
	if err := web.Decode(r, &upd); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	id := web.Param(r, "id")
	if err := h.Core.Update(ctx, claims, id, upd, v.Now); err != nil {
		switch validate.Cause(err) {
		case validate.ErrInvalidID:
			return webv1.NewRequestError(err, http.StatusBadRequest)
		case validate.ErrNotFound:
			return webv1.NewRequestError(err, http.StatusNotFound)
		case auth.ErrForbidden:
			return webv1.NewRequestError(err, http.StatusForbidden)
		default:
			return fmt.Errorf("ID[%s] User[%+v]: %w", id, &upd, err)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Delete removes a product from the system.
func (h Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return errors.New("claims missing from context")
	}

	id := web.Param(r, "id")
	if err := h.Core.Delete(ctx, claims, id); err != nil {
		switch validate.Cause(err) {
		case validate.ErrInvalidID:
			return webv1.NewRequestError(err, http.StatusBadRequest)
		default:
			return fmt.Errorf("ID[%s]: %w", id, err)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Query returns a list of products with paging.
func (h Handlers) Query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page := web.Param(r, "page")
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		return webv1.NewRequestError(fmt.Errorf("invalid page format, page[%s]", page), http.StatusBadRequest)
	}
	rows := web.Param(r, "rows")
	rowsPerPage, err := strconv.Atoi(rows)
	if err != nil {
		return webv1.NewRequestError(fmt.Errorf("invalid rows format, rows[%s]", rows), http.StatusBadRequest)
	}

	products, err := h.Core.Query(ctx, pageNumber, rowsPerPage)
	if err != nil {
		return fmt.Errorf("unable to query for products: %w", err)
	}

	return web.Respond(ctx, w, products, http.StatusOK)
}

// QueryByID returns a product by its ID.
func (h Handlers) QueryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := web.Param(r, "id")
	prod, err := h.Core.QueryByID(ctx, id)
	if err != nil {
		switch validate.Cause(err) {
		case validate.ErrInvalidID:
			return webv1.NewRequestError(err, http.StatusBadRequest)
		case validate.ErrNotFound:
			return webv1.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("ID[%s]: %w", id, err)
		}
	}

	return web.Respond(ctx, w, prod, http.StatusOK)
}
