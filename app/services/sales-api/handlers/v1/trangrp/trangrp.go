// Package trangrp maintains the group of handlers for transaction example.
package trangrp

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/sys/core"
	v1 "github.com/ardanlabs/service/business/web/v1"
	"github.com/ardanlabs/service/foundation/web"
)

// Handlers manages the set of product endpoints.
type Handlers struct {
	usrCore *user.Core
	prdCore *product.Core
}

// New constructs a handlers for route access.
func New(usrCore *user.Core, prdCore *product.Core) *Handlers {
	return &Handlers{
		usrCore: usrCore,
		prdCore: prdCore,
	}
}

// executeUnderTransaction constructs a new Handlers value with the core apis
// using a store transaction that was created via middleware.
func (h *Handlers) executeUnderTransaction(ctx context.Context) (*Handlers, error) {
	if tr, ok := core.GetTransaction(ctx); ok {
		usrCore, err := h.usrCore.ExecuteUnderTransaction(tr)
		if err != nil {
			return nil, err
		}

		prdCore, err := h.prdCore.ExecuteUnderTransaction(tr)
		if err != nil {
			return nil, err
		}

		h = &Handlers{
			usrCore: usrCore,
			prdCore: prdCore,
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
		return err
	}

	np, err := toCoreNewProduct(app.Product)
	if err != nil {
		return v1.NewRequestError(err, http.StatusBadRequest)
	}

	nu, err := toCoreNewUser(app.User)
	if err != nil {
		return v1.NewRequestError(err, http.StatusBadRequest)
	}

	usr, err := h.usrCore.Create(ctx, nu)
	if err != nil {
		return err
	}

	np.UserID = usr.ID

	prd, err := h.prdCore.Create(ctx, np)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, toAppProduct(prd), http.StatusCreated)
}
