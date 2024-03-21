package mid

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/business/core/crud/product"
	"github.com/ardanlabs/service/business/web/auth"
	"github.com/ardanlabs/service/business/web/errs"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/google/uuid"
)

type ctxProductKey int

const productKey ctxProductKey = 1

// GetProduct returns the product from the context.
func GetProduct(ctx context.Context) (product.Product, error) {
	v, ok := ctx.Value(productKey).(product.Product)
	if !ok {
		return product.Product{}, errors.New("product not found in context")
	}

	return v, nil
}

func setProduct(ctx context.Context, prd product.Product) context.Context {
	return context.WithValue(ctx, productKey, prd)
}

// AuthorizeProduct executes the specified role and extracts the specified
// product from the DB if a product id is specified in the call. Depending on
// the rule specified, the userid from the claims may be compared with the
// specified user id from the product.
func AuthorizeProduct(a *auth.Auth, prdCore *product.Core) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			var userID uuid.UUID

			if id := web.Param(r, "product_id"); id != "" {
				var err error
				productID, err := uuid.Parse(id)
				if err != nil {
					return errs.NewTrusted(ErrInvalidID, http.StatusBadRequest)
				}

				prd, err := prdCore.QueryByID(ctx, productID)
				if err != nil {
					switch {
					case errors.Is(err, product.ErrNotFound):
						return errs.NewTrusted(err, http.StatusNoContent)
					default:
						return fmt.Errorf("querybyid: productID[%s]: %w", productID, err)
					}
				}

				userID = prd.UserID
				ctx = setProduct(ctx, prd)
			}

			claims := getClaims(ctx)

			if err := a.Authorize(ctx, claims, userID, auth.RuleAdminOrSubject); err != nil {
				return auth.NewAuthError("authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", claims.Roles, auth.RuleAdminOrSubject, err)
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
