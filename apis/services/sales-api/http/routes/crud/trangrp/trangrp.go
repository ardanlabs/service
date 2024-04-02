// Package trangrp maintains the group of handlers for transaction example.
package trangrp

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/core/crud/tranapp"
	"github.com/ardanlabs/service/business/api/errs"
	"github.com/ardanlabs/service/foundation/web"
)

type handlers struct {
	tran *tranapp.Core
}

func new(tran *tranapp.Core) *handlers {
	return &handlers{
		tran: tran,
	}
}

func (h *handlers) create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app tranapp.NewTran
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	prd, err := h.tran.Create(ctx, app)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, prd, http.StatusCreated)
}
