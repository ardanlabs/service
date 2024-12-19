// Package vproductbus provides business access to view product domain.
package vproductbus

import (
	"context"
	"fmt"

	"github.com/ardanlabs/service/business/sdk/order"
	"github.com/ardanlabs/service/business/sdk/page"
	"github.com/ardanlabs/service/foundation/otel"
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Product, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// Business manages the set of APIs for view product access.
type Business struct {
	storer Storer
}

// NewBusiness constructs a vproduct business API for use.
func NewBusiness(storer Storer) *Business {
	return &Business{
		storer: storer,
	}
}

// Query retrieves a list of existing products.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Product, error) {
	ctx, span := otel.AddSpan(ctx, "business.vproductbus.query")
	defer span.End()

	users, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return users, nil
}

// Count returns the total number of products.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.vproductbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}
