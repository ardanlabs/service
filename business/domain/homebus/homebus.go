// Package homebus provides business access to home domain.
package homebus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/business/sdk/delegate"
	"github.com/ardanlabs/service/business/sdk/order"
	"github.com/ardanlabs/service/business/sdk/page"
	"github.com/ardanlabs/service/business/sdk/sqldb"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/otel"
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
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, hme Home) error
	Update(ctx context.Context, hme Home) error
	Delete(ctx context.Context, hme Home) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Home, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, homeID uuid.UUID) (Home, error)
	QueryByUserID(ctx context.Context, userID uuid.UUID) ([]Home, error)
}

// Business manages the set of APIs for home api access.
type Business struct {
	log      *logger.Logger
	userBus  *userbus.Business
	delegate *delegate.Delegate
	storer   Storer
}

// NewBusiness constructs a home business API for use.
func NewBusiness(log *logger.Logger, userBus *userbus.Business, delegate *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:      log,
		userBus:  userBus,
		delegate: delegate,
		storer:   storer,
	}
}

// NewWithTx constructs a new domain value that will use the
// specified transaction in any store related calls.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
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

// Create adds a new home to the system.
func (b *Business) Create(ctx context.Context, nh NewHome) (Home, error) {
	ctx, span := otel.AddSpan(ctx, "business.homebus.create")
	defer span.End()

	usr, err := b.userBus.QueryByID(ctx, nh.UserID)
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

	if err := b.storer.Create(ctx, hme); err != nil {
		return Home{}, fmt.Errorf("create: %w", err)
	}

	return hme, nil
}

// Update modifies information about a home.
func (b *Business) Update(ctx context.Context, hme Home, uh UpdateHome) (Home, error) {
	ctx, span := otel.AddSpan(ctx, "business.homebus.update")
	defer span.End()

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

	if err := b.storer.Update(ctx, hme); err != nil {
		return Home{}, fmt.Errorf("update: %w", err)
	}

	return hme, nil
}

// Delete removes the specified home.
func (b *Business) Delete(ctx context.Context, hme Home) error {
	ctx, span := otel.AddSpan(ctx, "business.homebus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, hme); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of existing homes.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Home, error) {
	ctx, span := otel.AddSpan(ctx, "business.homebus.query")
	defer span.End()

	hmes, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return hmes, nil
}

// Count returns the total number of homes.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.homebus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the home by the specified ID.
func (b *Business) QueryByID(ctx context.Context, homeID uuid.UUID) (Home, error) {
	ctx, span := otel.AddSpan(ctx, "business.homebus.querybyid")
	defer span.End()

	hme, err := b.storer.QueryByID(ctx, homeID)
	if err != nil {
		return Home{}, fmt.Errorf("query: homeID[%s]: %w", homeID, err)
	}

	return hme, nil
}

// QueryByUserID finds the homes by a specified User ID.
func (b *Business) QueryByUserID(ctx context.Context, userID uuid.UUID) ([]Home, error) {
	ctx, span := otel.AddSpan(ctx, "business.homebus.querybyuserid")
	defer span.End()

	hmes, err := b.storer.QueryByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return hmes, nil
}
