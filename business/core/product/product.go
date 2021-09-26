// Package product provides an example of a core business API. Right now these
// calls are just wrapping the data/store layer. But at some point you will
// want auditing or something that isn't specific to the data/store layer.
package product

import (
	"context"
	"fmt"
	"time"

	"github.com/ardanlabs/service/business/data/dbproduct"
	"github.com/ardanlabs/service/business/sys/auth"
	"github.com/ardanlabs/service/business/sys/validate"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Core manages the set of API's for product access.
type Core struct {
	log  *zap.SugaredLogger
	data dbproduct.Data
}

// NewCore constructs a core for product api access.
func NewCore(log *zap.SugaredLogger, db *sqlx.DB) Core {
	return Core{
		log:  log,
		data: dbproduct.NewData(log, db),
	}
}

// Create adds a Product to the database. It returns the created Product with
// fields like ID and DateCreated populated.
func (c Core) Create(ctx context.Context, np NewProduct, now time.Time) (Product, error) {
	if err := validate.Check(np); err != nil {
		return Product{}, fmt.Errorf("validating data: %w", err)
	}

	dbPrd, err := c.data.Create(ctx, toDBNewProduct(np), now)
	if err != nil {
		return Product{}, fmt.Errorf("create: %w", err)
	}

	return toProduct(dbPrd), nil
}

// Update modifies data about a Product. It will error if the specified ID is
// invalid or does not reference an existing Product.
func (c Core) Update(ctx context.Context, claims auth.Claims, productID string, up UpdateProduct, now time.Time) error {
	if err := validate.CheckID(productID); err != nil {
		return validate.ErrInvalidID
	}

	if err := validate.Check(up); err != nil {
		return fmt.Errorf("validating data: %w", err)
	}

	// If you are not an admin.
	if !claims.Authorized(auth.RoleAdmin) {
		return auth.ErrForbidden
	}

	if err := c.data.Update(ctx, productID, toDBUpdateProduct(up), now); err != nil {
		return fmt.Errorf("update: %w", err)
	}

	return nil
}

// Delete removes the product identified by a given ID.
func (c Core) Delete(ctx context.Context, claims auth.Claims, productID string) error {
	if err := validate.CheckID(productID); err != nil {
		return validate.ErrInvalidID
	}

	// If you are not an admin.
	if !claims.Authorized(auth.RoleAdmin) {
		return auth.ErrForbidden
	}

	if err := c.data.Delete(ctx, productID); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query gets all Products from the database.
func (c Core) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]Product, error) {
	dbPrds, err := c.data.Query(ctx, pageNumber, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return toProductSlice(dbPrds), nil
}

// QueryByID finds the product identified by a given ID.
func (c Core) QueryByID(ctx context.Context, productID string) (Product, error) {
	if err := validate.CheckID(productID); err != nil {
		return Product{}, validate.ErrInvalidID
	}

	dbPrd, err := c.data.QueryByID(ctx, productID)
	if err != nil {
		return Product{}, fmt.Errorf("query: %w", err)
	}

	return toProduct(dbPrd), nil
}

// QueryByUserID finds the product identified by a given User ID.
func (c Core) QueryByUserID(ctx context.Context, userID string) ([]Product, error) {
	if err := validate.CheckID(userID); err != nil {
		return nil, validate.ErrInvalidID
	}

	dbPrds, err := c.data.QueryByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return toProductSlice(dbPrds), nil
}
