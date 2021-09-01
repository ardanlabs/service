// Package product provides an example of a core business API. Right now these
// calls are just wrapping the data/store layer. But at some point you will
// want auditing or something that isn't specific to the data/store layer.
package product

import (
	"context"
	"fmt"
	"time"

	"github.com/ardanlabs/service/business/data/store/product"
	"github.com/ardanlabs/service/business/sys/auth"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Core manages the set of API's for product access.
type Core struct {
	log     *zap.SugaredLogger
	product product.Store
}

// NewCore constructs a core for product api access.
func NewCore(log *zap.SugaredLogger, db *sqlx.DB) Core {
	return Core{
		log:     log,
		product: product.NewStore(log, db),
	}
}

// Create adds a Product to the database. It returns the created Product with
// fields like ID and DateCreated populated.
func (c Core) Create(ctx context.Context, claims auth.Claims, np product.NewProduct, now time.Time) (product.Product, error) {

	// PERFORM PRE BUSINESS OPERATIONS

	prd, err := c.product.Create(ctx, claims, np, now)
	if err != nil {
		return product.Product{}, fmt.Errorf("create: %w", err)
	}

	// PERFORM POST BUSINESS OPERATIONS

	return prd, nil
}

// Update modifies data about a Product. It will error if the specified ID is
// invalid or does not reference an existing Product.
func (c Core) Update(ctx context.Context, claims auth.Claims, productID string, up product.UpdateProduct, now time.Time) error {

	// PERFORM PRE BUSINESS OPERATIONS

	if err := c.product.Update(ctx, claims, productID, up, now); err != nil {
		return fmt.Errorf("update: %w", err)
	}

	// PERFORM POST BUSINESS OPERATIONS

	return nil
}

// Delete removes the product identified by a given ID.
func (c Core) Delete(ctx context.Context, claims auth.Claims, productID string) error {

	// PERFORM PRE BUSINESS OPERATIONS

	if err := c.product.Delete(ctx, claims, productID); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// PERFORM POST BUSINESS OPERATIONS

	return nil
}

// Query gets all Products from the database.
func (c Core) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]product.Product, error) {

	// PERFORM PRE BUSINESS OPERATIONS

	products, err := c.product.Query(ctx, pageNumber, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	// PERFORM POST BUSINESS OPERATIONS

	return products, nil
}

// QueryByID finds the product identified by a given ID.
func (c Core) QueryByID(ctx context.Context, productID string) (product.Product, error) {

	// PERFORM PRE BUSINESS OPERATIONS

	prd, err := c.product.QueryByID(ctx, productID)
	if err != nil {
		return product.Product{}, fmt.Errorf("query: %w", err)
	}

	// PERFORM POST BUSINESS OPERATIONS

	return prd, nil
}

// QueryByUserID finds the product identified by a given User ID.
func (c Core) QueryByUserID(ctx context.Context, userID string) ([]product.Product, error) {

	// PERFORM PRE BUSINESS OPERATIONS

	products, err := c.product.QueryByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	// PERFORM POST BUSINESS OPERATIONS

	return products, nil
}
