package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/api/errs"
	"github.com/ardanlabs/service/business/api/mid"
	"github.com/ardanlabs/service/business/core/crud/home"
	"github.com/ardanlabs/service/business/core/crud/product"
	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/google/uuid"
)

// AuthorizeUser executes the specified role and extracts the specified user
// from the DB if a user id is specified in the call. Depending on the rule
// specified, the userid from the claims may be compared with the specified
// user id.
func AuthorizeUser(a *auth.Auth, userCore *user.Core, rule string) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			var userID uuid.UUID

			if id := web.Param(r, "user_id"); id != "" {
				var err error
				userID, err = uuid.Parse(id)
				if err != nil {
					return errs.New(errs.Unauthenticated, ErrInvalidID)
				}

				usr, err := userCore.QueryByID(ctx, userID)
				if err != nil {
					switch {
					case errors.Is(err, user.ErrNotFound):
						return errs.New(errs.Unauthenticated, err)
					default:
						return errs.Newf(errs.Unauthenticated, "querybyid: userID[%s]: %s", userID, err)
					}
				}

				ctx = mid.SetUser(ctx, usr)
			}

			claims := mid.GetClaims(ctx)
			if err := a.Authorize(ctx, claims, userID, rule); err != nil {
				return errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", claims.Roles, rule, err)
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

// AuthorizeProduct executes the specified role and extracts the specified
// product from the DB if a product id is specified in the call. Depending on
// the rule specified, the userid from the claims may be compared with the
// specified user id from the product.
func AuthorizeProduct(a *auth.Auth, productCore *product.Core) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			var userID uuid.UUID

			if id := web.Param(r, "product_id"); id != "" {
				var err error
				productID, err := uuid.Parse(id)
				if err != nil {
					return errs.New(errs.Unauthenticated, ErrInvalidID)
				}

				prd, err := productCore.QueryByID(ctx, productID)
				if err != nil {
					switch {
					case errors.Is(err, product.ErrNotFound):
						return errs.New(errs.Unauthenticated, err)
					default:
						return errs.Newf(errs.Internal, "querybyid: productID[%s]: %s", productID, err)
					}
				}

				userID = prd.UserID
				ctx = mid.SetProduct(ctx, prd)
			}

			claims := mid.GetClaims(ctx)

			if err := a.Authorize(ctx, claims, userID, auth.RuleAdminOrSubject); err != nil {
				return errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", claims.Roles, auth.RuleAdminOrSubject, err)
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

// AuthorizeHome executes the specified role and extracts the specified
// home from the DB if a home id is specified in the call. Depending on
// the rule specified, the userid from the claims may be compared with the
// specified user id from the home.
func AuthorizeHome(a *auth.Auth, homeCore *home.Core) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			var userID uuid.UUID

			if id := web.Param(r, "home_id"); id != "" {
				var err error
				homeID, err := uuid.Parse(id)
				if err != nil {
					return errs.New(errs.Unauthenticated, ErrInvalidID)
				}

				hme, err := homeCore.QueryByID(ctx, homeID)
				if err != nil {
					switch {
					case errors.Is(err, home.ErrNotFound):
						return errs.New(errs.Unauthenticated, err)
					default:
						return errs.Newf(errs.Unauthenticated, "querybyid: homeID[%s]: %s", homeID, err)
					}
				}

				userID = hme.UserID
				ctx = mid.SetHome(ctx, hme)
			}

			claims := mid.GetClaims(ctx)
			if err := a.Authorize(ctx, claims, userID, auth.RuleAdminOrSubject); err != nil {
				return errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", claims.Roles, auth.RuleAdminOrSubject, err)
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
