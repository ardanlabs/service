// Package reports provides an example of a business/service level API.
package reports

import (
	"context"
	"fmt"

	"github.com/ardanlabs/service/business/data/store/product"
	"github.com/ardanlabs/service/business/data/store/user"
	"github.com/ardanlabs/service/business/sys/auth"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Service manages the set of API's for payment functionality.
type Service struct {
	User    user.Store
	Product product.Store
}

// NewService constructs a payment service for api access.
func NewService(log *zap.SugaredLogger, db *sqlx.DB) Service {
	return Service{
		User:    user.NewStore(log, db),
		Product: product.NewStore(log, db),
	}
}

// UserProducts validates the user exists and returns products they have created.
func (s Service) UserProducts(ctx context.Context, claims auth.Claims, userID string) ([]product.Product, error) {
	user, err := s.User.QueryByID(ctx, claims, userID)
	if err != nil {
		return nil, fmt.Errorf("query user UserID[%s]: %w", user, err)
	}

	products, err := s.Product.QueryByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("query products UserID[%s]: %w", user, err)
	}

	return products, nil
}
