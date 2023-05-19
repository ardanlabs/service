// Package product provides an example of a core business API. Right now these
// calls are just wrapping the data/store layer. But at some point you will
// want auditing or something that isn't specific to the data/store layer.
package product

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ardanlabs/service/business/core/event"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/order"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound    = errors.New("product not found")
	ErrInvalidUser = errors.New("user not valid")
)

// =============================================================================

// Storer interface declares the behavior this package needs to perists and
// retrieve data.
type Storer interface {
	Create(ctx context.Context, prd Product) error
	Update(ctx context.Context, prd Product) error
	Delete(ctx context.Context, prd Product) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]Product, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, productID uuid.UUID) (Product, error)
	QueryByUserID(ctx context.Context, userID uuid.UUID) ([]Product, error)
}

// UserCore interface declares the behavior this package needs from the user
// core domain.
type UserCore interface {
	QueryByID(ctx context.Context, userID uuid.UUID) (user.User, error)
}

// =============================================================================

// Core manages the set of APIs for product access.
type Core struct {
	log     *zap.SugaredLogger
	evnCore *event.Core
	usrCore UserCore
	storer  Storer
}

// NewCore constructs a core for product api access.
func NewCore(log *zap.SugaredLogger, evnCore *event.Core, usrCore UserCore, storer Storer) *Core {
	core := Core{
		log:     log,
		evnCore: evnCore,
		usrCore: usrCore,
		storer:  storer,
	}

	core.registerEventHandlers(evnCore)

	return &core
}

// Create adds a Product to the database. It returns the created Product with
// fields like ID and DateCreated populated.
func (c *Core) Create(ctx context.Context, np NewProduct) (Product, error) {
	usr, err := c.usrCore.QueryByID(ctx, np.UserID)
	if err != nil {
		return Product{}, fmt.Errorf("user.querybyid: %s: %w", np.UserID, err)
	}

	if !usr.Enabled {
		return Product{}, ErrInvalidUser
	}

	now := time.Now()

	prd := Product{
		ID:          uuid.New(),
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
func (c *Core) Update(ctx context.Context, prd Product, up UpdateProduct) (Product, error) {
	if up.Name != nil {
		prd.Name = *up.Name
	}
	if up.Cost != nil {
		prd.Cost = *up.Cost
	}
	if up.Quantity != nil {
		prd.Quantity = *up.Quantity
	}
	prd.DateUpdated = time.Now()

	if err := c.storer.Update(ctx, prd); err != nil {
		return Product{}, fmt.Errorf("update: %w", err)
	}

	return prd, nil
}

// Delete removes the product identified by a given ID.
func (c *Core) Delete(ctx context.Context, prd Product) error {
	if err := c.storer.Delete(ctx, prd); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query gets all Products from the database.
func (c *Core) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]Product, error) {
	prds, err := c.storer.Query(ctx, filter, orderBy, pageNumber, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return prds, nil
}

// Count returns the total number of products in the store.
func (c *Core) Count(ctx context.Context, filter QueryFilter) (int, error) {
	return c.storer.Count(ctx, filter)
}

// QueryByID finds the product identified by a given ID.
func (c *Core) QueryByID(ctx context.Context, productID uuid.UUID) (Product, error) {
	prd, err := c.storer.QueryByID(ctx, productID)
	if err != nil {
		return Product{}, fmt.Errorf("query: productID[%s]: %w", productID, err)
	}

	return prd, nil
}

// QueryByUserID finds the products identified by a given User ID.
func (c *Core) QueryByUserID(ctx context.Context, userID uuid.UUID) ([]Product, error) {
	prds, err := c.storer.QueryByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return prds, nil
}
