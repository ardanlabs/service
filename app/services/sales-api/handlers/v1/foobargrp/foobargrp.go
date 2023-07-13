package foobargrp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/business/core/foobar"
	"github.com/ardanlabs/service/business/sys/core"
	"github.com/ardanlabs/service/business/web/auth"
	v1 "github.com/ardanlabs/service/business/web/v1"
	"github.com/ardanlabs/service/foundation/web"
)

// Handlers manages the set of product endpoints.
type Handlers struct {
	foobar *foobar.Core
	auth   *auth.Auth
}

// New constructs a handlers for route access.
func New(foobar *foobar.Core, auth *auth.Auth) *Handlers {
	return &Handlers{
		foobar: foobar,
		auth:   auth,
	}
}

func (h *Handlers) InTran(ctx context.Context) (*Handlers, error) {
	if tr, ok := core.GetTransactor(ctx); ok {
		fb, err := h.foobar.InTran(tr)
		if err != nil {
			return nil, err
		}
		return &Handlers{
			foobar: fb,
			auth:   h.auth,
		}, nil
	}
	return h, nil
}

// Create adds a new product to the system.
func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var err error
	h, err = h.InTran(ctx)
	if err != nil {
		return err
	}
	var app AppNewFoobar
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
	prd, err := h.foobar.Create(ctx, np, nu)

	if err != nil {
		return fmt.Errorf("create: app[%+v]: %w", app, err)
	}

	return web.Respond(ctx, w, toAppProduct(prd), http.StatusCreated)
}
