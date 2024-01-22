// Package usersummarygrp maintains the group of handlers for user summary access.
package usersummarygrp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/business/core/usersummary"
	v1 "github.com/ardanlabs/service/business/web/v1"
	"github.com/ardanlabs/service/business/web/v1/page"
	"github.com/ardanlabs/service/foundation/web"
)

// Handlers manages the set of user endpoints.
type Handlers struct {
	summary *usersummary.Core
}

func new(summary *usersummary.Core) *Handlers {
	return &Handlers{
		summary: summary,
	}
}

// Query returns a list of user summary data with paging.
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

	smms, err := h.summary.Query(ctx, filter, orderBy, page.Number, page.RowsPerPage)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	total, err := h.summary.Count(ctx, filter)
	if err != nil {
		return fmt.Errorf("count: %w", err)
	}

	return web.Respond(ctx, w, v1.NewPageDocument(toAppUsersSummary(smms), total, page.Number, page.RowsPerPage), http.StatusOK)
}
