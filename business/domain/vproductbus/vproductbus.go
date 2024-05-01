// Package vproductbus provides business access to view product domain.
package vproductbus

import (
	"context"
	"fmt"

	"github.com/ardanlabs/service/business/api/order"
)

// Storer interface declares the behavior this package needs to perists and
// retrieve data.
type Storer interface {
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]Product, error)
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
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]Product, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}

	users, err := b.storer.Query(ctx, filter, orderBy, pageNumber, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return users, nil
}

// Count returns the total number of products.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	if err := filter.Validate(); err != nil {
		return 0, err
	}

	return b.storer.Count(ctx, filter)
}
