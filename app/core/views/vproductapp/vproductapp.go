// Package vproductapp maintains the app layer api for the vproduct domain.
package vproductapp

import (
	"context"

	"github.com/ardanlabs/service/business/api/errs"
	"github.com/ardanlabs/service/business/api/page"
	"github.com/ardanlabs/service/business/core/views/vproductbus"
)

// Core manages the set of app layer api functions for the view product domain.
type Core struct {
	vproductBus *vproductbus.Core
}

// NewCore constructs a view product core API for use.
func NewCore(vproductBus *vproductbus.Core) *Core {
	return &Core{
		vproductBus: vproductBus,
	}
}

// Query returns a list of products with paging.
func (c *Core) Query(ctx context.Context, qp QueryParams) (page.Document[Product], error) {
	if err := validatePaging(qp); err != nil {
		return page.Document[Product]{}, err
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return page.Document[Product]{}, err
	}

	orderBy, err := parseOrder(qp)
	if err != nil {
		return page.Document[Product]{}, err
	}

	prds, err := c.vproductBus.Query(ctx, filter, orderBy, qp.Page, qp.Rows)
	if err != nil {
		return page.Document[Product]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := c.vproductBus.Count(ctx, filter)
	if err != nil {
		return page.Document[Product]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return page.NewDocument(toAppProducts(prds), total, qp.Page, qp.Rows), nil
}
