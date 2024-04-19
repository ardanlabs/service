// Package mid provides app level middleware support.
package mid

import (
	"context"
	"errors"

	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/domain/homebus"
	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/google/uuid"
)

// Handler represents the handler function that needs to be called.
type Handler func(context.Context) error

type ctxKey int

const (
	claimKey ctxKey = iota + 1
	userIDKey
	userKey
	productKey
	homeKey
)

func SetClaims(ctx context.Context, claims auth.Claims) context.Context {
	return context.WithValue(ctx, claimKey, claims)
}

func GetClaims(ctx context.Context) auth.Claims {
	v, ok := ctx.Value(claimKey).(auth.Claims)
	if !ok {
		return auth.Claims{}
	}
	return v
}

// GetUserID returns the claims from the context.
func GetUserID(ctx context.Context) (uuid.UUID, error) {
	v, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}, errors.New("user id not found in context")
	}

	return v, nil
}

// GetUser returns the user from the context.
func GetUser(ctx context.Context) (userbus.User, error) {
	v, ok := ctx.Value(userKey).(userbus.User)
	if !ok {
		return userbus.User{}, errors.New("user not found in context")
	}

	return v, nil
}

func SetUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func SetUser(ctx context.Context, usr userbus.User) context.Context {
	return context.WithValue(ctx, userKey, usr)
}

// GetProduct returns the product from the context.
func GetProduct(ctx context.Context) (productbus.Product, error) {
	v, ok := ctx.Value(productKey).(productbus.Product)
	if !ok {
		return productbus.Product{}, errors.New("product not found in context")
	}

	return v, nil
}

func SetProduct(ctx context.Context, prd productbus.Product) context.Context {
	return context.WithValue(ctx, productKey, prd)
}

// GetHome returns the home from the context.
func GetHome(ctx context.Context) (homebus.Home, error) {
	v, ok := ctx.Value(homeKey).(homebus.Home)
	if !ok {
		return homebus.Home{}, errors.New("home not found in context")
	}

	return v, nil
}

func SetHome(ctx context.Context, hme homebus.Home) context.Context {
	return context.WithValue(ctx, homeKey, hme)
}
