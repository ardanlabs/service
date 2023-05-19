// Package summary provides an example of a core business API that
// is based on a view.
package summary

import (
	"context"
	"fmt"

	"github.com/ardanlabs/service/business/data/order"
)

// Storer interface declares the behavior this package needs to perists and
// retrieve data.
type Storer interface {
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]Summary, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// =============================================================================

// Core manages the set of APIs for user access.
type Core struct {
	storer Storer
}

// NewCore constructs a core for use product api access.
func NewCore(storer Storer) *Core {
	return &Core{
		storer: storer,
	}
}

// Query retrieves a list of existing users from the database.
func (c *Core) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]Summary, error) {
	users, err := c.storer.Query(ctx, filter, orderBy, pageNumber, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return users, nil
}

// Count returns the total number of users in the store.
func (c *Core) Count(ctx context.Context, filter QueryFilter) (int, error) {
	return c.storer.Count(ctx, filter)
}
