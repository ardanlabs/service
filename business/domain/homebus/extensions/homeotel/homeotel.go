// Package homeotel provides an extension for homebus that adds
// otel tracking.
package homeotel

import (
	"context"

	"github.com/ardanlabs/service/business/domain/homebus"
	"github.com/ardanlabs/service/business/sdk/order"
	"github.com/ardanlabs/service/business/sdk/page"
	"github.com/ardanlabs/service/business/sdk/sqldb"
	"github.com/ardanlabs/service/foundation/otel"
	"github.com/google/uuid"
)

// Extension provides a wrapper for otel functionality around the homebus.
type Extension struct {
	bus homebus.ExtBusiness
}

// NewExtension constructs a new extension that wraps the homebus with otel.
func NewExtension() homebus.Extension {
	return func(bus homebus.ExtBusiness) homebus.ExtBusiness {
		return &Extension{
			bus: bus,
		}
	}
}

// NewWithTx does not apply otel.
func (ext *Extension) NewWithTx(tx sqldb.CommitRollbacker) (homebus.ExtBusiness, error) {
	return ext.bus.NewWithTx(tx)
}

// Create applies otel to the home creation process.
func (ext *Extension) Create(ctx context.Context, nh homebus.NewHome) (homebus.Home, error) {
	ctx, span := otel.AddSpan(ctx, "business.homebus.create")
	defer span.End()

	usr, err := ext.bus.Create(ctx, nh)
	if err != nil {
		return homebus.Home{}, err
	}

	return usr, nil
}

// Update applies otel to the home update process.
func (ext *Extension) Update(ctx context.Context, hme homebus.Home, uh homebus.UpdateHome) (homebus.Home, error) {
	ctx, span := otel.AddSpan(ctx, "business.homebus.update")
	defer span.End()

	usr, err := ext.bus.Update(ctx, hme, uh)
	if err != nil {
		return homebus.Home{}, err
	}

	return usr, nil
}

// Delete applies otel to the home delete process.
func (ext *Extension) Delete(ctx context.Context, hme homebus.Home) error {
	ctx, span := otel.AddSpan(ctx, "business.homebus.delete")
	defer span.End()

	if err := ext.bus.Delete(ctx, hme); err != nil {
		return err
	}

	return nil
}

// Query applies otel to the home query process.
func (ext *Extension) Query(ctx context.Context, filter homebus.QueryFilter, orderBy order.By, page page.Page) ([]homebus.Home, error) {
	ctx, span := otel.AddSpan(ctx, "business.homebus.query")
	defer span.End()

	usr, err := ext.bus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

// Count applies otel to the home count process.
func (ext *Extension) Count(ctx context.Context, filter homebus.QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.homebus.count")
	defer span.End()

	usr, err := ext.bus.Count(ctx, filter)
	if err != nil {
		return 0, err
	}

	return usr, nil
}

// QueryByID applies otel to the home query by id process.
func (ext *Extension) QueryByID(ctx context.Context, homeID uuid.UUID) (homebus.Home, error) {
	ctx, span := otel.AddSpan(ctx, "business.homebus.queryByID")
	defer span.End()

	usr, err := ext.bus.QueryByID(ctx, homeID)
	if err != nil {
		return homebus.Home{}, err
	}

	return usr, nil
}

// QueryByUserID applies otel to the home query by user id process.
func (ext *Extension) QueryByUserID(ctx context.Context, userID uuid.UUID) ([]homebus.Home, error) {
	ctx, span := otel.AddSpan(ctx, "business.homebus.queryByUserID")
	defer span.End()

	usr, err := ext.bus.QueryByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return usr, nil
}
