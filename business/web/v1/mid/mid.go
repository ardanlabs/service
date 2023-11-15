// Package mid contains the set of middleware functions.
package mid

import (
	"context"

	"github.com/ardanlabs/service/business/core/home"
	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/web/v1/auth"
	"github.com/google/uuid"
)

// ctxKey represents the type of value for the context key.
type ctxKey int

const (
	claimKey ctxKey = iota + 1
	userIDKey
	userKey
	prodKey
	homeKey
)

// =============================================================================

// GetUserID returns the claims from the context.
func GetUserID(ctx context.Context) uuid.UUID {
	v, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}
	}
	return v
}

// GetUser returns the user from the context.
func GetUser(ctx context.Context) user.User {
	v, ok := ctx.Value(userKey).(user.User)
	if !ok {
		return user.User{}
	}
	return v
}

// GetProduct returns the product from the context.
func GetProduct(ctx context.Context) product.Product {
	v, ok := ctx.Value(prodKey).(product.Product)
	if !ok {
		return product.Product{}
	}
	return v
}

// GetHome returns the home from the context.
func GetHome(ctx context.Context) home.Home {
	v, ok := ctx.Value(homeKey).(home.Home)
	if !ok {
		return home.Home{}
	}
	return v
}

// =============================================================================

func setClaims(ctx context.Context, claims auth.Claims) context.Context {
	return context.WithValue(ctx, claimKey, claims)
}

func getClaims(ctx context.Context) auth.Claims {
	v, ok := ctx.Value(claimKey).(auth.Claims)
	if !ok {
		return auth.Claims{}
	}
	return v
}

func setUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func setUser(ctx context.Context, usr user.User) context.Context {
	return context.WithValue(ctx, userKey, usr)
}

func setProduct(ctx context.Context, prd product.Product) context.Context {
	return context.WithValue(ctx, prodKey, prd)
}

func setHome(ctx context.Context, hme home.Home) context.Context {
	return context.WithValue(ctx, homeKey, hme)
}
