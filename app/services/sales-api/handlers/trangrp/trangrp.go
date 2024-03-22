// Package trangrp maintains the group of handlers for transaction example.
package trangrp

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/business/core/crud/product"
	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/web/errs"
	"github.com/ardanlabs/service/foundation/web"
)

type handlers struct {
	user    *user.Core
	product *product.Core
}

func new(user *user.Core, product *product.Core) *handlers {
	return &handlers{
		user:    user,
		product: product,
	}
}

// create adds a new user and product at the same time under a single transaction.
func (h *handlers) create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return err
	}

	var app AppNewTran
	if err := web.Decode(r, &app); err != nil {
		return errs.NewTrusted(err, http.StatusBadRequest)
	}

	np, err := toCoreNewProduct(app.Product)
	if err != nil {
		return errs.NewTrusted(err, http.StatusBadRequest)
	}

	nu, err := toCoreNewUser(app.User)
	if err != nil {
		return errs.NewTrusted(err, http.StatusBadRequest)
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
