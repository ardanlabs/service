// Package vproductapi maintains the group of handlers for detailed product data.
package vproductapi

import (
	"context"

	"github.com/ardanlabs/service/business/api/errs"
	"github.com/ardanlabs/service/business/api/page"
	"github.com/ardanlabs/service/business/core/views/vproduct"
)

// Handlers manages the set of handler functions for this domain.
type Handlers struct {
	vproduct *vproduct.Core
}

// New constructs a Handlers for use.
func New(vproduct *vproduct.Core) *Handlers {
	return &Handlers{
		vproduct: vproduct,
	}
}

// Query returns a list of products with paging.
func (h *Handlers) Query(ctx context.Context, qp QueryParams) (page.Document[AppProduct], error) {
	if err := validatePaging(qp); err != nil {
		return page.Document[AppProduct]{}, err
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return page.Document[AppProduct]{}, err
	}

	orderBy, err := parseOrder(qp)
	if err != nil {
		return page.Document[AppProduct]{}, err
	}

	prds, err := h.vproduct.Query(ctx, filter, orderBy, qp.Page, qp.Rows)
	if err != nil {
		return page.Document[AppProduct]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := h.vproduct.Count(ctx, filter)
	if err != nil {
		return page.Document[AppProduct]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return page.NewDocument(toAppProducts(prds), total, qp.Page, qp.Rows), nil
}
