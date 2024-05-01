// Package productbus provides business access to product domain.
package productbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ardanlabs/service/business/api/delegate"
	"github.com/ardanlabs/service/business/api/order"
	"github.com/ardanlabs/service/business/api/transaction"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/google/uuid"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound     = errors.New("product not found")
	ErrUserDisabled = errors.New("user disabled")
	ErrInvalidCost  = errors.New("cost not valid")
)

// Storer interface declares the behavior this package needs to perists and
// retrieve data.
type Storer interface {
	NewWithTx(tx transaction.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, prd Product) error
	Update(ctx context.Context, prd Product) error
	Delete(ctx context.Context, prd Product) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]Product, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, productID uuid.UUID) (Product, error)
	QueryByUserID(ctx context.Context, userID uuid.UUID) ([]Product, error)
}

// Business manages the set of APIs for product access.
type Business struct {
	log      *logger.Logger
	userBus  *userbus.Business
	delegate *delegate.Delegate
	storer   Storer
}

// NewBusiness constructs a product business API for use.
func NewBusiness(log *logger.Logger, userBus *userbus.Business, delegate *delegate.Delegate, storer Storer) *Business {
	b := Business{
		log:      log,
		userBus:  userBus,
		delegate: delegate,
		storer:   storer,
	}

	b.registerDelegateFunctions()

	return &b
}

// NewWithTx constructs a new business value that will use the
// specified transaction in any store related calls.
func (b *Business) NewWithTx(tx transaction.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	userBus, err := b.userBus.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	bus := Business{
		log:      b.log,
		userBus:  userBus,
		delegate: b.delegate,
		storer:   storer,
	}

	return &bus, nil
}

// Create adds a new product to the system.
func (b *Business) Create(ctx context.Context, np NewProduct) (Product, error) {
	usr, err := b.userBus.QueryByID(ctx, np.UserID)
	if err != nil {
		return Product{}, fmt.Errorf("user.querybyid: %s: %w", np.UserID, err)
	}

	if np.Cost < 0 {
		return Product{}, ErrInvalidCost
	}

	if !usr.Enabled {
		return Product{}, ErrUserDisabled
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

	if err := b.storer.Create(ctx, prd); err != nil {
		return Product{}, fmt.Errorf("create: %w", err)
	}

	return prd, nil
}

// Update modifies information about a product.
func (b *Business) Update(ctx context.Context, prd Product, up UpdateProduct) (Product, error) {
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

	if err := b.storer.Update(ctx, prd); err != nil {
		return Product{}, fmt.Errorf("update: %w", err)
	}

	return prd, nil
}

// Delete removes the specified product.
func (b *Business) Delete(ctx context.Context, prd Product) error {
	if err := b.storer.Delete(ctx, prd); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of existing products.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]Product, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}

	prds, err := b.storer.Query(ctx, filter, orderBy, pageNumber, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return prds, nil
}

// Count returns the total number of products.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	if err := filter.Validate(); err != nil {
		return 0, err
	}

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the product by the specified Ib.
func (b *Business) QueryByID(ctx context.Context, productID uuid.UUID) (Product, error) {
	prd, err := b.storer.QueryByID(ctx, productID)
	if err != nil {
		return Product{}, fmt.Errorf("query: productID[%s]: %w", productID, err)
	}

	return prd, nil
}

// QueryByUserID finds the products by a specified User Ib.
func (b *Business) QueryByUserID(ctx context.Context, userID uuid.UUID) ([]Product, error) {
	prds, err := b.storer.QueryByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return prds, nil
}
