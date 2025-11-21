// Package auditotel provides an extension for auditbus that adds
// otel tracking.
package auditotel

import (
	"context"

	"github.com/ardanlabs/service/business/domain/auditbus"
	"github.com/ardanlabs/service/business/sdk/order"
	"github.com/ardanlabs/service/business/sdk/page"
	"github.com/ardanlabs/service/foundation/otel"
)

// Extension provides a wrapper for otel functionality around the auditbus.
type Extension struct {
	bus auditbus.ExtBusiness
}

// NewExtension constructs a new extension that wraps the auditbus with otel.
func NewExtension() auditbus.Extension {
	return func(bus auditbus.ExtBusiness) auditbus.ExtBusiness {
		return &Extension{
			bus: bus,
		}
	}
}

// Create adds a new audit record to the system.
func (ext *Extension) Create(ctx context.Context, na auditbus.NewAudit) (auditbus.Audit, error) {
	ctx, span := otel.AddSpan(ctx, "business.auditbus.create")
	defer span.End()

	usr, err := ext.bus.Create(ctx, na)
	if err != nil {
		return auditbus.Audit{}, err
	}

	return usr, nil
}

// Query queries the audit records.
func (ext *Extension) Query(ctx context.Context, filter auditbus.QueryFilter, orderBy order.By, page page.Page) ([]auditbus.Audit, error) {
	ctx, span := otel.AddSpan(ctx, "business.auditbus.query")
	defer span.End()

	usr, err := ext.bus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

// Count counts the number of audit records.
func (ext *Extension) Count(ctx context.Context, filter auditbus.QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.auditbus.count")
	defer span.End()

	usr, err := ext.bus.Count(ctx, filter)
	if err != nil {
		return 0, err
	}

	return usr, nil
}
