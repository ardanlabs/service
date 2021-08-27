// Package report provides an example of a core business API.
package report

import (
	"context"
	"fmt"

	"github.com/ardanlabs/service/business/data/store/product"
	"github.com/ardanlabs/service/business/data/store/user"
	"github.com/ardanlabs/service/business/sys/auth"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Report manages the set of API's for report functionality.
type Report struct {
	User    user.Store
	Product product.Store
}

// NewReport constructs a Report for api access.
func NewReport(log *zap.SugaredLogger, db *sqlx.DB) Report {
	return Report{
		User:    user.NewStore(log, db),
		Product: product.NewStore(log, db),
	}
}

// UserProducts validates the user exists and returns products they have created.
func (r Report) UserProducts(ctx context.Context, claims auth.Claims, userID string) ([]product.Product, error) {
	user, err := r.User.QueryByID(ctx, claims, userID)
	if err != nil {
		return nil, fmt.Errorf("query user UserID[%s]: %w", user, err)
	}

	products, err := r.Product.QueryByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("query products UserID[%s]: %w", user, err)
	}

	return products, nil
}
