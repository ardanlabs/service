// Package productapp maintains the app layer api for the product domain.
package productapp

import (
	"context"
	"errors"

	"github.com/ardanlabs/service/business/api/errs"
	"github.com/ardanlabs/service/business/api/mid"
	"github.com/ardanlabs/service/business/api/page"
	"github.com/ardanlabs/service/business/core/crud/product"
)

// Set of error variables for handling product group errors.
var (
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// Core manages the set of handler functions for this domain.
type Core struct {
	product *product.Core
}

// New constructs a Handlers for use.
func New(product *product.Core) *Core {
	return &Core{
		product: product,
	}
}

// Create adds a new product to the system.
func (c *Core) Create(ctx context.Context, app NewProduct) (Product, error) {
	np, err := toBusNewProduct(ctx, app)
	if err != nil {
		return Product{}, errs.New(errs.FailedPrecondition, err)
	}

	prd, err := c.product.Create(ctx, np)
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

	updPrd, err := c.product.Update(ctx, prd, toBusUpdateProduct(app))
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

	if err := c.product.Delete(ctx, prd); err != nil {
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

	prds, err := c.product.Query(ctx, filter, orderBy, qp.Page, qp.Rows)
	if err != nil {
		return page.Document[Product]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := c.product.Count(ctx, filter)
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
