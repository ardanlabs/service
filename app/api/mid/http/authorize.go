package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/app/api/mid"
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/core/crud/homebus"
	"github.com/ardanlabs/service/business/core/crud/productbus"
	"github.com/ardanlabs/service/business/core/crud/userbus"
	"github.com/ardanlabs/service/foundation/authapi"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/google/uuid"
)

// ErrInvalidID represents a condition where the id is not a uuid.
var ErrInvalidID = errors.New("ID is not in its proper form")

// Authorize executes the specified role and does not extract any domain data.
func Authorize(log *logger.Logger, authAPI *authapi.AuthAPI, rule string) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			userID, err := mid.GetUserID(ctx)
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			authInfo := authapi.AuthInfo{
				Claims: mid.GetClaims(ctx),
				UserID: userID,
				Rule:   rule,
			}

			if err := authAPI.Authorize(ctx, authInfo); err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

// AuthorizeUser executes the specified role and extracts the specified user
// from the DB if a user id is specified in the call. Depending on the rule
// specified, the userid from the claims may be compared with the specified
// user id.
func AuthorizeUser(log *logger.Logger, authAPI *authapi.AuthAPI, userBus *userbus.Core, rule string) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			var userID uuid.UUID

			if id := web.Param(r, "user_id"); id != "" {
				var err error
				userID, err = uuid.Parse(id)
				if err != nil {
					return errs.New(errs.Unauthenticated, ErrInvalidID)
				}

				usr, err := userBus.QueryByID(ctx, userID)
				if err != nil {
					switch {
					case errors.Is(err, userbus.ErrNotFound):
						return errs.New(errs.Unauthenticated, err)
					default:
						return errs.Newf(errs.Unauthenticated, "querybyid: userID[%s]: %s", userID, err)
					}
				}

				ctx = mid.SetUser(ctx, usr)
			}

			authInfo := authapi.AuthInfo{
				Claims: mid.GetClaims(ctx),
				UserID: userID,
				Rule:   rule,
			}

			if err := authAPI.Authorize(ctx, authInfo); err != nil {
				return errs.New(errs.Unauthenticated, err)
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
func AuthorizeProduct(log *logger.Logger, authAPI *authapi.AuthAPI, productBus *productbus.Core) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			var userID uuid.UUID

			if id := web.Param(r, "product_id"); id != "" {
				var err error
				productID, err := uuid.Parse(id)
				if err != nil {
					return errs.New(errs.Unauthenticated, ErrInvalidID)
				}

				prd, err := productBus.QueryByID(ctx, productID)
				if err != nil {
					switch {
					case errors.Is(err, productbus.ErrNotFound):
						return errs.New(errs.Unauthenticated, err)
					default:
						return errs.Newf(errs.Internal, "querybyid: productID[%s]: %s", productID, err)
					}
				}

				userID = prd.UserID
				ctx = mid.SetProduct(ctx, prd)
			}

			authInfo := authapi.AuthInfo{
				Claims: mid.GetClaims(ctx),
				UserID: userID,
				Rule:   auth.RuleAdminOrSubject,
			}

			if err := authAPI.Authorize(ctx, authInfo); err != nil {
				return errs.New(errs.Unauthenticated, err)
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
func AuthorizeHome(log *logger.Logger, authAPI *authapi.AuthAPI, homeBus *homebus.Core) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			var userID uuid.UUID

			if id := web.Param(r, "home_id"); id != "" {
				var err error
				homeID, err := uuid.Parse(id)
				if err != nil {
					return errs.New(errs.Unauthenticated, ErrInvalidID)
				}

				hme, err := homeBus.QueryByID(ctx, homeID)
				if err != nil {
					switch {
					case errors.Is(err, homebus.ErrNotFound):
						return errs.New(errs.Unauthenticated, err)
					default:
						return errs.Newf(errs.Unauthenticated, "querybyid: homeID[%s]: %s", homeID, err)
					}
				}

				userID = hme.UserID
				ctx = mid.SetHome(ctx, hme)
			}

			authInfo := authapi.AuthInfo{
				Claims: mid.GetClaims(ctx),
				UserID: userID,
				Rule:   auth.RuleAdminOrSubject,
			}

			if err := authAPI.Authorize(ctx, authInfo); err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
