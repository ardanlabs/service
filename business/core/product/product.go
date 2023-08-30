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
	"github.com/ardanlabs/service/business/data/transaction"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/google/uuid"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound    = errors.New("product not found")
	ErrInvalidUser = errors.New("user not valid")
	ErrInvalidCost = errors.New("cost not valid")
)

// =============================================================================

// Storer interface declares the behavior this package needs to perists and
// retrieve data.
type Storer interface {
	ExecuteUnderTransaction(tx transaction.Transaction) (Storer, error)
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
	ExecuteUnderTransaction(tx transaction.Transaction) (*user.Core, error)
	QueryByID(ctx context.Context, userID uuid.UUID) (user.User, error)
}

// =============================================================================

// Core manages the set of APIs for product access.
type Core struct {
	log     *logger.Logger
	evnCore *event.Core
	usrCore UserCore
	storer  Storer
}

// NewCore constructs a core for product api access.
func NewCore(log *logger.Logger, evnCore *event.Core, usrCore UserCore, storer Storer) *Core {
	c := Core{
		log:     log,
		evnCore: evnCore,
		usrCore: usrCore,
		storer:  storer,
	}

	c.registerEventHandlers(evnCore)

	return &c
}

// ExecuteUnderTransaction constructs a new Core value that will use the
// specified transaction in any store related calls.
func (c *Core) ExecuteUnderTransaction(tx transaction.Transaction) (*Core, error) {
	storer, err := c.storer.ExecuteUnderTransaction(tx)
	if err != nil {
		return nil, err
	}

	usrCore, err := c.usrCore.ExecuteUnderTransaction(tx)
	if err != nil {
		return nil, err
	}

	c = &Core{
		storer:  storer,
		evnCore: c.evnCore,
		usrCore: usrCore,
		log:     c.log,
	}

	return c, nil
}

// Create adds a new product to the system.
func (c *Core) Create(ctx context.Context, np NewProduct) (Product, error) {
	usr, err := c.usrCore.QueryByID(ctx, np.UserID)
	if err != nil {
		return Product{}, fmt.Errorf("user.querybyid: %s: %w", np.UserID, err)
	}

	if np.Cost < 0 {
		return Product{}, ErrInvalidCost
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

// Update modifies information about a product.
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

// Delete removes the specified product.
func (c *Core) Delete(ctx context.Context, prd Product) error {
	if err := c.storer.Delete(ctx, prd); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of existing products.
func (c *Core) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]Product, error) {
	prds, err := c.storer.Query(ctx, filter, orderBy, pageNumber, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return prds, nil
}

// Count returns the total number of products.
func (c *Core) Count(ctx context.Context, filter QueryFilter) (int, error) {
	return c.storer.Count(ctx, filter)
}

// QueryByID finds the product by the specified ID.
func (c *Core) QueryByID(ctx context.Context, productID uuid.UUID) (Product, error) {
	prd, err := c.storer.QueryByID(ctx, productID)
	if err != nil {
		return Product{}, fmt.Errorf("query: productID[%s]: %w", productID, err)
	}

	return prd, nil
}

// QueryByUserID finds the products by a specified User ID.
func (c *Core) QueryByUserID(ctx context.Context, userID uuid.UUID) ([]Product, error) {
	prds, err := c.storer.QueryByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return prds, nil
}
