// Package report provides an example of a business package that brings
// together multiple core business packages.
package report

import (
	"context"
	"fmt"

	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/user"
)

// Core manages the set apis for report functionality.
type Core struct {
	User    user.Core
	Product product.Core
}

// NewCore constructs a core for report api access.
func NewCore(user user.Core, product product.Core) Core {
	return Core{
		User:    user,
		Product: product,
	}
}

// UserProducts validates the user exists and returns products they have created.
func (c Core) UserProducts(ctx context.Context, userID string) ([]product.Product, error) {
	if _, err := c.User.QueryByID(ctx, userID); err != nil {
		return nil, fmt.Errorf("query user UserID[%s]: %w", userID, err)
	}

	products, err := c.Product.QueryByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query products UserID[%s]: %w", userID, err)
	}

	return products, nil
}
