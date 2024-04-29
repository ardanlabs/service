// Package vproductapp maintains the app layer api for the vproduct domain.
package vproductapp

import (
	"context"

	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/app/api/page"
	"github.com/ardanlabs/service/business/domain/vproductbus"
)

// App manages the set of app layer api functions for the view product domain.
type App struct {
	vproductBus *vproductbus.Business
}

// NewApp constructs a view product app API for use.
func NewApp(vproductBus *vproductbus.Business) *App {
	return &App{
		vproductBus: vproductBus,
	}
}

// Query returns a list of products with paging.
func (a *App) Query(ctx context.Context, qp QueryParams) (page.Document[Product], error) {
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

	prds, err := a.vproductBus.Query(ctx, filter, orderBy, qp.Page, qp.Rows)
	if err != nil {
		return page.Document[Product]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.vproductBus.Count(ctx, filter)
	if err != nil {
		return page.Document[Product]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return page.NewDocument(toAppProducts(prds), total, qp.Page, qp.Rows), nil
}
