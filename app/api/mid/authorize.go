package mid

import (
	"context"
	"errors"

	"github.com/ardanlabs/service/app/api/auth"
	"github.com/ardanlabs/service/app/api/authclient"
	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/business/domain/homebus"
	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/google/uuid"
)

// ErrInvalidID represents a condition where the id is not a uuid.
var ErrInvalidID = errors.New("ID is not in its proper form")

// Authorize validates authorization via the auth service.
func Authorize(ctx context.Context, log *logger.Logger, client *authclient.Client, rule string, next Handler) (Encoder, error) {
	userID, err := GetUserID(ctx)
	if err != nil {
		return nil, errs.New(errs.Unauthenticated, err)
	}

	auth := authclient.Authorize{
		Claims: GetClaims(ctx),
		UserID: userID,
		Rule:   rule,
	}

	if err := client.Authorize(ctx, auth); err != nil {
		return nil, errs.New(errs.Unauthenticated, err)
	}

	return next(ctx)
}

// AuthorizeUser executes the specified role and extracts the specified
// user from the DB if a user id is specified in the call. Depending on the rule
// specified, the userid from the claims may be compared with the specified
// user id.
func AuthorizeUser(ctx context.Context, log *logger.Logger, client *authclient.Client, userBus *userbus.Business, rule string, id string, next Handler) (Encoder, error) {
	var userID uuid.UUID

	if id != "" {
		var err error
		userID, err = uuid.Parse(id)
		if err != nil {
			return nil, errs.New(errs.Unauthenticated, ErrInvalidID)
		}

		usr, err := userBus.QueryByID(ctx, userID)
		if err != nil {
			switch {
			case errors.Is(err, userbus.ErrNotFound):
				return nil, errs.New(errs.Unauthenticated, err)
			default:
				return nil, errs.Newf(errs.Unauthenticated, "querybyid: userID[%s]: %s", userID, err)
			}
		}

		ctx = setUser(ctx, usr)
	}

	auth := authclient.Authorize{
		Claims: GetClaims(ctx),
		UserID: userID,
		Rule:   rule,
	}

	if err := client.Authorize(ctx, auth); err != nil {
		return nil, errs.New(errs.Unauthenticated, err)
	}

	return next(ctx)
}

// AuthorizeProduct executes the specified role and extracts the specified
// product from the DB if a product id is specified in the call. Depending on
// the rule specified, the userid from the claims may be compared with the
// specified user id from the product.
func AuthorizeProduct(ctx context.Context, log *logger.Logger, client *authclient.Client, productBus *productbus.Business, id string, next Handler) (Encoder, error) {
	var userID uuid.UUID

	if id != "" {
		var err error
		productID, err := uuid.Parse(id)
		if err != nil {
			return nil, errs.New(errs.Unauthenticated, ErrInvalidID)
		}

		prd, err := productBus.QueryByID(ctx, productID)
		if err != nil {
			switch {
			case errors.Is(err, productbus.ErrNotFound):
				return nil, errs.New(errs.Unauthenticated, err)
			default:
				return nil, errs.Newf(errs.Internal, "querybyid: productID[%s]: %s", productID, err)
			}
		}

		userID = prd.UserID
		ctx = setProduct(ctx, prd)
	}

	auth := authclient.Authorize{
		UserID: userID,
		Claims: GetClaims(ctx),
		Rule:   auth.RuleAdminOrSubject,
	}

	if err := client.Authorize(ctx, auth); err != nil {
		return nil, errs.New(errs.Unauthenticated, err)
	}

	return next(ctx)
}

// AuthorizeHome executes the specified role and extracts the specified
// home from the DB if a home id is specified in the call. Depending on
// the rule specified, the userid from the claims may be compared with the
// specified user id from the home.
func AuthorizeHome(ctx context.Context, log *logger.Logger, client *authclient.Client, homeBus *homebus.Business, id string, next Handler) (Encoder, error) {
	var userID uuid.UUID

	if id != "" {
		var err error
		homeID, err := uuid.Parse(id)
		if err != nil {
			return nil, errs.New(errs.Unauthenticated, ErrInvalidID)
		}

		hme, err := homeBus.QueryByID(ctx, homeID)
		if err != nil {
			switch {
			case errors.Is(err, homebus.ErrNotFound):
				return nil, errs.New(errs.Unauthenticated, err)
			default:
				return nil, errs.Newf(errs.Unauthenticated, "querybyid: homeID[%s]: %s", homeID, err)
			}
		}

		userID = hme.UserID
		ctx = setHome(ctx, hme)
	}

	auth := authclient.Authorize{
		Claims: GetClaims(ctx),
		UserID: userID,
		Rule:   auth.RuleAdminOrSubject,
	}

	if err := client.Authorize(ctx, auth); err != nil {
		return nil, errs.New(errs.Unauthenticated, err)
	}

	return next(ctx)
}
