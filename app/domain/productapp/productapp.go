// Package productapp maintains the app layer api for the product domain.
package productapp

import (
	"context"

	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/app/api/mid"
	"github.com/ardanlabs/service/app/api/page"
	"github.com/ardanlabs/service/business/domain/productbus"
)

// Core manages the set of app layer api functions for the product domain.
type Core struct {
	productBus *productbus.Core
}

// NewCore constructs a product core API for use.
func NewCore(productBus *productbus.Core) *Core {
	return &Core{
		productBus: productBus,
	}
}

// Create adds a new product to the system.
func (c *Core) Create(ctx context.Context, app NewProduct) (Product, error) {
	np, err := toBusNewProduct(ctx, app)
	if err != nil {
		return Product{}, errs.New(errs.FailedPrecondition, err)
	}

	prd, err := c.productBus.Create(ctx, np)
	if err != nil {
		return Product{}, errs.Newf(errs.Internal, "create: prd[%+v]: %s", prd, err)
	}

	return toAppProduct(prd), nil
}

// Update updates an existing product.
func (c *Core) Update(ctx context.Context, app UpdateProduct) (Product, error) {
	prd, err := mid.GetProduct(ctx)
	if err != nil {
		return Product{}, errs.Newf(errs.Internal, "product missing in context: %s", err)
	}

	updPrd, err := c.productBus.Update(ctx, prd, toBusUpdateProduct(app))
	if err != nil {
		return Product{}, errs.Newf(errs.Internal, "update: productID[%s] up[%+v]: %s", prd.ID, app, err)
	}

	return toAppProduct(updPrd), nil
}

// Delete removes a product from the system.
func (c *Core) Delete(ctx context.Context) error {
	prd, err := mid.GetProduct(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "productID missing in context: %s", err)
	}

	if err := c.productBus.Delete(ctx, prd); err != nil {
		return errs.Newf(errs.Internal, "delete: productID[%s]: %s", prd.ID, err)
	}

	return nil
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

	prds, err := c.productBus.Query(ctx, filter, orderBy, qp.Page, qp.Rows)
	if err != nil {
		return page.Document[Product]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := c.productBus.Count(ctx, filter)
	if err != nil {
		return page.Document[Product]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return page.NewDocument(toAppProducts(prds), total, qp.Page, qp.Rows), nil
}

// QueryByID returns a product by its ID.
func (c *Core) QueryByID(ctx context.Context) (Product, error) {
	prd, err := mid.GetProduct(ctx)
	if err != nil {
		return Product{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return toAppProduct(prd), nil
}
