// Package trangrp maintains the group of handlers for transaction example.
package trangrp

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/transaction"
	v1 "github.com/ardanlabs/service/business/web/v1"
	"github.com/ardanlabs/service/foundation/web"
)

// Handlers manages the set of product endpoints.
type Handlers struct {
	user    *user.Core
	product *product.Core
}

// New constructs a handlers for route access.
func New(user *user.Core, product *product.Core) *Handlers {
	return &Handlers{
		user:    user,
		product: product,
	}
}

// executeUnderTransaction constructs a new Handlers value with the core apis
// using a store transaction that was created via middleware.
func (h *Handlers) executeUnderTransaction(ctx context.Context) (*Handlers, error) {
	if tx, ok := transaction.Get(ctx); ok {
		user, err := h.user.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		product, err := h.product.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		h = &Handlers{
			user:    user,
			product: product,
		}

		return h, nil
	}

	return h, nil
}

// Create adds a new user and product at the same time under a single transaction.
func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return err
	}

	var app AppNewTran
	if err := web.Decode(r, &app); err != nil {
		return v1.NewRequestError(err, http.StatusBadRequest)
	}

	np, err := toCoreNewProduct(app.Product)
	if err != nil {
		return v1.NewRequestError(err, http.StatusBadRequest)
	}

	nu, err := toCoreNewUser(app.User)
	if err != nil {
		return v1.NewRequestError(err, http.StatusBadRequest)
	}

	usr, err := h.user.Create(ctx, nu)
	if err != nil {
		return err
	}

	np.UserID = usr.ID

	prd, err := h.product.Create(ctx, np)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, toAppProduct(prd), http.StatusCreated)
}
