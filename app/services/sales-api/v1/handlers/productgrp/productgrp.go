// Package productgrp maintains the group of handlers for product access.
package productgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/page"
	"github.com/ardanlabs/service/business/web/v1/mid"
	"github.com/ardanlabs/service/business/web/v1/response"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/google/uuid"
)

// Set of error variables for handling product group errors.
var (
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// Handlers manages the set of product endpoints.
type Handlers struct {
	product *product.Core
	user    *user.Core
}

// New constructs a handlers for route access.
func New(product *product.Core, user *user.Core) *Handlers {
	return &Handlers{
		product: product,
		user:    user,
	}
}

// Create adds a new product to the system.
func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app AppNewProduct
	if err := web.Decode(r, &app); err != nil {
		return response.NewError(err, http.StatusBadRequest)
	}

	np, err := toCoreNewProduct(ctx, app)
	if err != nil {
		return response.NewError(err, http.StatusBadRequest)
	}

	prd, err := h.product.Create(ctx, np)
	if err != nil {
		return fmt.Errorf("create: app[%+v]: %w", app, err)
	}

	return web.Respond(ctx, w, toAppProduct(prd), http.StatusCreated)
}

// Update updates a product in the system.
func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app AppUpdateProduct
	if err := web.Decode(r, &app); err != nil {
		return response.NewError(err, http.StatusBadRequest)
	}

	prd := mid.GetProduct(ctx)

	prd, err := h.product.Update(ctx, prd, toCoreUpdateProduct(app))
	if err != nil {
		return fmt.Errorf("update: productID[%s] app[%+v]: %w", prd.ID, app, err)
	}

	return web.Respond(ctx, w, toAppProduct(prd), http.StatusOK)
}

// Delete removes a product from the system.
func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	prd := mid.GetProduct(ctx)

	if err := h.product.Delete(ctx, prd); err != nil {
		return fmt.Errorf("delete: productID[%s]: %w", prd.ID, err)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Query returns a list of products with paging.
func (h *Handlers) Query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page, err := page.Parse(r)
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

	prds, err := h.product.Query(ctx, filter, orderBy, page.Number, page.RowsPerPage)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	// -------------------------------------------------------------------------
	// Capture the unique set of users

	users := make(map[uuid.UUID]user.User)
	if len(prds) > 0 {
		for _, prd := range prds {
			users[prd.UserID] = user.User{}
		}

		userIDs := make([]uuid.UUID, 0, len(users))
		for userID := range users {
			userIDs = append(userIDs, userID)
		}

		usrs, err := h.user.QueryByIDs(ctx, userIDs)
		if err != nil {
			return fmt.Errorf("user.querybyids: userIDs[%s]: %w", userIDs, err)
		}

		for _, usr := range usrs {
			users[usr.ID] = usr
		}
	}

	// -------------------------------------------------------------------------

	total, err := h.product.Count(ctx, filter)
	if err != nil {
		return fmt.Errorf("count: %w", err)
	}

	return web.Respond(ctx, w, response.NewPageDocument(toAppProductsDetails(prds, users), total, page.Number, page.RowsPerPage), http.StatusOK)
}

// QueryByID returns a product by its ID.
func (h *Handlers) QueryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return web.Respond(ctx, w, toAppProduct(mid.GetProduct(ctx)), http.StatusOK)
}
