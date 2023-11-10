package mid

import (
	"context"

	"github.com/ardanlabs/service/business/core/home"
	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/user"
)

type ctxKey int

const userKey ctxKey = 1

const prodKey ctxKey = 2

const homeKey ctxKey = 3

// =============================================================================

// SetUser stores the user in the context.
func SetUser(ctx context.Context, usr user.User) context.Context {
	return context.WithValue(ctx, userKey, usr)
}

// GetUser returns the user from the context.
func GetUser(ctx context.Context) user.User {
	v, ok := ctx.Value(userKey).(user.User)
	if !ok {
		return user.User{}
	}
	return v
}

// SetProduct stores the product in the context.
func SetProduct(ctx context.Context, prd product.Product) context.Context {
	return context.WithValue(ctx, prodKey, prd)
}

// GetProduct returns the product from the context.
func GetProduct(ctx context.Context) product.Product {
	v, ok := ctx.Value(prodKey).(product.Product)
	if !ok {
		return product.Product{}
	}
	return v
}

// SetHome stores the home in the context.
func SetHome(ctx context.Context, hme home.Home) context.Context {
	return context.WithValue(ctx, homeKey, hme)
}

// GetHome returns the home from the context.
func GetHome(ctx context.Context) home.Home {
	v, ok := ctx.Value(homeKey).(home.Home)
	if !ok {
		return home.Home{}
	}
	return v
}
