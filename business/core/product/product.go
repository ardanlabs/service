// Package product provides an example of a core business API. Right now these
// calls are just wrapping the data/store layer. But at some point you will
// want auditing or something that isn't specific to the data/store layer.
package product

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ardanlabs/service/business/data/sort"
	"github.com/ardanlabs/service/business/sys/validate"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound     = errors.New("product not found")
	ErrInvalidID    = errors.New("ID is not in its proper form")
	ErrInvalidOrder = errors.New("validating order by")
)

// Storer interface declares the behavior this package needs to perists and
// retrieve data.
type Storer interface {
	Create(ctx context.Context, prd Product) error
	Update(ctx context.Context, prd Product) error
	Delete(ctx context.Context, productID string) error
	Query(ctx context.Context, orderBy sort.OrderBy, pageNumber int, rowsPerPage int) ([]Product, error)
	QueryByID(ctx context.Context, productID string) (Product, error)
	QueryByUserID(ctx context.Context, userID string) ([]Product, error)
}

// Core manages the set of APIs for product access.
type Core struct {
	storer Storer
}

// NewCore constructs a core for product api access.
func NewCore(storer Storer) *Core {
	return &Core{
		storer: storer,
	}
}

// Create adds a Product to the database. It returns the created Product with
// fields like ID and DateCreated populated.
func (c *Core) Create(ctx context.Context, np NewProduct, now time.Time) (Product, error) {
	if err := validate.Check(np); err != nil {
		return Product{}, fmt.Errorf("validating data: %w", err)
	}

	prd := Product{
		ID:          validate.GenerateID(),
		Name:        np.Name,
		Cost:        np.Cost,
		Quantity:    np.Quantity,
		UserID:      np.UserID,
		DateCreated: now,
		DateUpdated: now,
	}

	if err := c.storer.Create(ctx, prd); err != nil {
		return Product{}, fmt.Errorf("create: %w", err)
	}

	return prd, nil
}

// Update modifies data about a Product. It will error if the specified ID is
// invalid or does not reference an existing Product.
func (c *Core) Update(ctx context.Context, productID string, up UpdateProduct, now time.Time) error {
	if err := validate.CheckID(productID); err != nil {
		return ErrInvalidID
	}

	if err := validate.Check(up); err != nil {
		return fmt.Errorf("validating data: %w", err)
	}

	prd, err := c.storer.QueryByID(ctx, productID)
	if err != nil {
		return fmt.Errorf("updating product productID[%s]: %w", productID, err)
	}

	if up.Name != nil {
		prd.Name = *up.Name
	}
	if up.Cost != nil {
		prd.Cost = *up.Cost
	}
	if up.Quantity != nil {
		prd.Quantity = *up.Quantity
	}
	prd.DateUpdated = now

	if err := c.storer.Update(ctx, prd); err != nil {
		return fmt.Errorf("update: %w", err)
	}

	return nil
}

// Delete removes the product identified by a given ID.
func (c *Core) Delete(ctx context.Context, productID string) error {
	if err := validate.CheckID(productID); err != nil {
		return ErrInvalidID
	}

	if err := c.storer.Delete(ctx, productID); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query gets all Products from the database.
func (c *Core) Query(ctx context.Context, orderBy sort.OrderBy, pageNumber int, rowsPerPage int) ([]Product, error) {
	if err := Order.Check(&orderBy); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidOrder, err.Error())
	}

	prds, err := c.storer.Query(ctx, orderBy, pageNumber, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return prds, nil
}

// QueryByID finds the product identified by a given ID.
func (c *Core) QueryByID(ctx context.Context, productID string) (Product, error) {
	if err := validate.CheckID(productID); err != nil {
		return Product{}, ErrInvalidID
	}

	prd, err := c.storer.QueryByID(ctx, productID)
	if err != nil {
		return Product{}, fmt.Errorf("query: %w", err)
	}

	return prd, nil
}

// QueryByUserID finds the products identified by a given User ID.
func (c *Core) QueryByUserID(ctx context.Context, userID string) ([]Product, error) {
	if err := validate.CheckID(userID); err != nil {
		return nil, ErrInvalidID
	}

	prds, err := c.storer.QueryByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return prds, nil
}
