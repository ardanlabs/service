// Package vproductapp maintains the app layer api for the vproduct domain.
package vproductapp

import (
	"context"

	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/app/sdk/page"
	"github.com/ardanlabs/service/business/domain/vproductbus"
	"github.com/ardanlabs/service/business/sdk/order"
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
	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return page.Document[Product]{}, err
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return page.Document[Product]{}, err
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return page.Document[Product]{}, err
	}

	prds, err := a.vproductBus.Query(ctx, filter, orderBy, pg.Number, pg.RowsPerPage)
	if err != nil {
		return page.Document[Product]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.vproductBus.Count(ctx, filter)
	if err != nil {
		return page.Document[Product]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return page.NewDocument(toAppProducts(prds), total, pg.Number, pg.RowsPerPage), nil
}
