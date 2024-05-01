// Package mid provides app level middleware support.
package mid

import (
	"context"
	"errors"

	"github.com/ardanlabs/service/app/api/auth"
	"github.com/ardanlabs/service/business/api/transaction"
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
	trKey
)

func setClaims(ctx context.Context, claims auth.Claims) context.Context {
	return context.WithValue(ctx, claimKey, claims)
}

// GetClaims returns the claims from the context.
func GetClaims(ctx context.Context) auth.Claims {
	v, ok := ctx.Value(claimKey).(auth.Claims)
	if !ok {
		return auth.Claims{}
	}
	return v
}

func setUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID returns the user id from the context.
func GetUserID(ctx context.Context) (uuid.UUID, error) {
	v, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}, errors.New("user id not found in context")
	}

	return v, nil
}

func setUser(ctx context.Context, usr userbus.User) context.Context {
	return context.WithValue(ctx, userKey, usr)
}

// GetUser returns the user from the context.
func GetUser(ctx context.Context) (userbus.User, error) {
	v, ok := ctx.Value(userKey).(userbus.User)
	if !ok {
		return userbus.User{}, errors.New("user not found in context")
	}

	return v, nil
}

func setProduct(ctx context.Context, prd productbus.Product) context.Context {
	return context.WithValue(ctx, productKey, prd)
}

// GetProduct returns the product from the context.
func GetProduct(ctx context.Context) (productbus.Product, error) {
	v, ok := ctx.Value(productKey).(productbus.Product)
	if !ok {
		return productbus.Product{}, errors.New("product not found in context")
	}

	return v, nil
}

func setHome(ctx context.Context, hme homebus.Home) context.Context {
	return context.WithValue(ctx, homeKey, hme)
}

// GetHome returns the home from the context.
func GetHome(ctx context.Context) (homebus.Home, error) {
	v, ok := ctx.Value(homeKey).(homebus.Home)
	if !ok {
		return homebus.Home{}, errors.New("home not found in context")
	}

	return v, nil
}

func setTran(ctx context.Context, tx transaction.CommitRollbacker) context.Context {
	return context.WithValue(ctx, trKey, tx)
}

// GetTran retrieves the value that can manage a transaction.
func GetTran(ctx context.Context) (transaction.CommitRollbacker, bool) {
	v, ok := ctx.Value(trKey).(transaction.CommitRollbacker)
	return v, ok
}
