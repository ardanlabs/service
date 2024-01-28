package mid

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/business/core/product"
	v1 "github.com/ardanlabs/service/business/web/v1"
	"github.com/ardanlabs/service/business/web/v1/auth"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/google/uuid"
)

type ctxProductKey int

const productKey ctxProductKey = 1

// GetProduct returns the product from the context.
func GetProduct(ctx context.Context) product.Product {
	v, ok := ctx.Value(productKey).(product.Product)
	if !ok {
		return product.Product{}
	}
	return v
}

func setProduct(ctx context.Context, prd product.Product) context.Context {
	return context.WithValue(ctx, productKey, prd)
}

// AuthorizeProduct executes the specified role and extracts the specified
// product from the DB if a product id is specified in the call. Depending on
// the rule specified, the userid from the claims may be compared with the
// specified user id from the product.
func AuthorizeProduct(a *auth.Auth, rule string, prdCore *product.Core) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			var userID uuid.UUID

			if id := web.Param(r, "product_id"); id != "" {
				var err error
				productID, err := uuid.Parse(id)
				if err != nil {
					return v1.NewTrustedError(ErrInvalidID, http.StatusBadRequest)
				}

				prd, err := prdCore.QueryByID(ctx, productID)
				if err != nil {
					switch {
					case errors.Is(err, product.ErrNotFound):
						return v1.NewTrustedError(err, http.StatusNoContent)
					default:
						return fmt.Errorf("querybyid: productID[%s]: %w", productID, err)
					}
				}

				userID = prd.UserID
				ctx = setProduct(ctx, prd)
			}

			claims := getClaims(ctx)

			if err := a.Authorize(ctx, claims, userID, rule); err != nil {
				return auth.NewAuthError("authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", claims.Roles, rule, err)
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
