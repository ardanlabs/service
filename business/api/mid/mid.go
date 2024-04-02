// Package mid contains the set of values the middleware is responsible
// to extract and set.
package mid

import (
	"context"
	"errors"

	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/core/crud/home"
	"github.com/ardanlabs/service/business/core/crud/product"
	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/google/uuid"
)

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
func GetUser(ctx context.Context) (user.User, error) {
	v, ok := ctx.Value(userKey).(user.User)
	if !ok {
		return user.User{}, errors.New("user not found in context")
	}

	return v, nil
}

func SetUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func SetUser(ctx context.Context, usr user.User) context.Context {
	return context.WithValue(ctx, userKey, usr)
}

// GetProduct returns the product from the context.
func GetProduct(ctx context.Context) (product.Product, error) {
	v, ok := ctx.Value(productKey).(product.Product)
	if !ok {
		return product.Product{}, errors.New("product not found in context")
	}

	return v, nil
}

func SetProduct(ctx context.Context, prd product.Product) context.Context {
	return context.WithValue(ctx, productKey, prd)
}

// GetHome returns the home from the context.
func GetHome(ctx context.Context) (home.Home, error) {
	v, ok := ctx.Value(homeKey).(home.Home)
	if !ok {
		return home.Home{}, errors.New("home not found in context")
	}

	return v, nil
}

func SetHome(ctx context.Context, hme home.Home) context.Context {
	return context.WithValue(ctx, homeKey, hme)
}
