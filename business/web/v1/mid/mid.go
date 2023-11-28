// Package mid contains the set of middleware functions.
package mid

import (
	"context"

	"github.com/ardanlabs/service/business/web/v1/auth"
)

type ctxKey int

const claimKey ctxKey = 1

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
