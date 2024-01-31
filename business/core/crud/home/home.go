// Package home provides a business access to home data in the system.
package home

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ardanlabs/service/business/core/crud/delegate"
	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/data/transaction"
	"github.com/ardanlabs/service/business/web/v1/order"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/google/uuid"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound     = errors.New("home not found")
	ErrUserDisabled = errors.New("user disabled")
)

// Storer interface declares the behaviour this package needs to persist and
// retrieve data.
type Storer interface {
	ExecuteUnderTransaction(tx transaction.Transaction) (Storer, error)
	Create(ctx context.Context, hme Home) error
	Update(ctx context.Context, hme Home) error
	Delete(ctx context.Context, hme Home) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]Home, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, homeID uuid.UUID) (Home, error)
	QueryByUserID(ctx context.Context, userID uuid.UUID) ([]Home, error)
}

// Core manages the set of APIs for home api access.
type Core struct {
	log      *logger.Logger
	usrCore  *user.Core
	delegate *delegate.Delegate
	storer   Storer
}

// NewCore constructs a home core API for use.
func NewCore(log *logger.Logger, usrCore *user.Core, delegate *delegate.Delegate, storer Storer) *Core {
	return &Core{
		log:      log,
		usrCore:  usrCore,
		delegate: delegate,
		storer:   storer,
	}
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

	core := Core{
		log:      c.log,
		usrCore:  usrCore,
		delegate: c.delegate,
		storer:   storer,
	}

	return &core, nil
}

// Create adds a new home to the system.
func (c *Core) Create(ctx context.Context, nh NewHome) (Home, error) {
	usr, err := c.usrCore.QueryByID(ctx, nh.UserID)
	if err != nil {
		return Home{}, fmt.Errorf("user.querybyid: %s: %w", nh.UserID, err)
	}

	if !usr.Enabled {
		return Home{}, ErrUserDisabled
	}

	now := time.Now()

	hme := Home{
		ID:   uuid.New(),
		Type: nh.Type,
		Address: Address{
			Address1: nh.Address.Address1,
			Address2: nh.Address.Address2,
			ZipCode:  nh.Address.ZipCode,
			City:     nh.Address.City,
			State:    nh.Address.State,
			Country:  nh.Address.Country,
		},
		UserID:      nh.UserID,
		DateCreated: now,
		DateUpdated: now,
	}

	if err := c.storer.Create(ctx, hme); err != nil {
		return Home{}, fmt.Errorf("create: %w", err)
	}

	return hme, nil
}

// Update modifies information about a home.
func (c *Core) Update(ctx context.Context, hme Home, uh UpdateHome) (Home, error) {
	if uh.Type != nil {
		hme.Type = *uh.Type
	}

	if uh.Address != nil {
		if uh.Address.Address1 != nil {
			hme.Address.Address1 = *uh.Address.Address1
		}

		if uh.Address.Address2 != nil {
			hme.Address.Address2 = *uh.Address.Address2
		}

		if uh.Address.ZipCode != nil {
			hme.Address.ZipCode = *uh.Address.ZipCode
		}

		if uh.Address.City != nil {
			hme.Address.City = *uh.Address.City
		}

		if uh.Address.State != nil {
			hme.Address.State = *uh.Address.State
		}

		if uh.Address.Country != nil {
			hme.Address.Country = *uh.Address.Country
		}
	}

	hme.DateUpdated = time.Now()

	if err := c.storer.Update(ctx, hme); err != nil {
		return Home{}, fmt.Errorf("update: %w", err)
	}

	return hme, nil
}

// Delete removes the specified home.
func (c *Core) Delete(ctx context.Context, hme Home) error {
	if err := c.storer.Delete(ctx, hme); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of existing homes.
func (c *Core) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]Home, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}

	hmes, err := c.storer.Query(ctx, filter, orderBy, pageNumber, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return hmes, nil
}

// Count returns the total number of homes.
func (c *Core) Count(ctx context.Context, filter QueryFilter) (int, error) {
	if err := filter.Validate(); err != nil {
		return 0, err
	}

	return c.storer.Count(ctx, filter)
}

// QueryByID finds the home by the specified ID.
func (c *Core) QueryByID(ctx context.Context, homeID uuid.UUID) (Home, error) {
	hme, err := c.storer.QueryByID(ctx, homeID)
	if err != nil {
		return Home{}, fmt.Errorf("query: homeID[%s]: %w", homeID, err)
	}

	return hme, nil
}

// QueryByUserID finds the homes by a specified User ID.
func (c *Core) QueryByUserID(ctx context.Context, userID uuid.UUID) ([]Home, error) {
	hmes, err := c.storer.QueryByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return hmes, nil
}
