package mid

import (
	"context"
	"errors"

	"github.com/ardanlabs/service/apis/api/authclient"
	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/domain/homebus"
	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/google/uuid"
)

// ErrInvalidID represents a condition where the id is not a uuid.
var ErrInvalidID = errors.New("ID is not in its proper form")

// Authorize executes the specified role and does not extract any domain data.
func Authorize(ctx context.Context, log *logger.Logger, client *authclient.Client, rule string) error {
	userID, err := GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	auth := authclient.Authorize{
		Claims: GetClaims(ctx),
		UserID: userID,
		Rule:   rule,
	}

	if err := client.Authorize(ctx, auth); err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	return nil
}

// AuthorizeUser executes the specified role and extracts the specified user
// from the DB if a user id is specified in the call. Depending on the rule
// specified, the userid from the claims may be compared with the specified
// user id.
func AuthorizeUser(ctx context.Context, log *logger.Logger, client *authclient.Client, userBus *userbus.Core, rule string, id string) (context.Context, error) {
	var userID uuid.UUID

	if id != "" {
		var err error
		userID, err = uuid.Parse(id)
		if err != nil {
			return ctx, errs.New(errs.Unauthenticated, ErrInvalidID)
		}

		usr, err := userBus.QueryByID(ctx, userID)
		if err != nil {
			switch {
			case errors.Is(err, userbus.ErrNotFound):
				return ctx, errs.New(errs.Unauthenticated, err)
			default:
				return ctx, errs.Newf(errs.Unauthenticated, "querybyid: userID[%s]: %s", userID, err)
			}
		}

		ctx = SetUser(ctx, usr)
	}

	auth := authclient.Authorize{
		Claims: GetClaims(ctx),
		UserID: userID,
		Rule:   rule,
	}

	if err := client.Authorize(ctx, auth); err != nil {
		return ctx, errs.New(errs.Unauthenticated, err)
	}

	return ctx, nil
}

// AuthorizeProduct executes the specified role and extracts the specified
// product from the DB if a product id is specified in the call. Depending on
// the rule specified, the userid from the claims may be compared with the
// specified user id from the product.
func AuthorizeProduct(ctx context.Context, log *logger.Logger, client *authclient.Client, productBus *productbus.Core, id string) (context.Context, error) {
	var userID uuid.UUID

	if id != "" {
		var err error
		productID, err := uuid.Parse(id)
		if err != nil {
			return ctx, errs.New(errs.Unauthenticated, ErrInvalidID)
		}

		prd, err := productBus.QueryByID(ctx, productID)
		if err != nil {
			switch {
			case errors.Is(err, productbus.ErrNotFound):
				return ctx, errs.New(errs.Unauthenticated, err)
			default:
				return ctx, errs.Newf(errs.Internal, "querybyid: productID[%s]: %s", productID, err)
			}
		}

		userID = prd.UserID
		ctx = SetProduct(ctx, prd)
	}

	auth := authclient.Authorize{
		Claims: GetClaims(ctx),
		UserID: userID,
		Rule:   auth.RuleAdminOrSubject,
	}

	if err := client.Authorize(ctx, auth); err != nil {
		return ctx, errs.New(errs.Unauthenticated, err)
	}

	return ctx, nil
}

// AuthorizeHome executes the specified role and extracts the specified
// home from the DB if a home id is specified in the call. Depending on
// the rule specified, the userid from the claims may be compared with the
// specified user id from the home.
func AuthorizeHome(ctx context.Context, log *logger.Logger, client *authclient.Client, homeBus *homebus.Core, id string) (context.Context, error) {
	var userID uuid.UUID

	if id != "" {
		var err error
		homeID, err := uuid.Parse(id)
		if err != nil {
			return ctx, errs.New(errs.Unauthenticated, ErrInvalidID)
		}

		hme, err := homeBus.QueryByID(ctx, homeID)
		if err != nil {
			switch {
			case errors.Is(err, homebus.ErrNotFound):
				return ctx, errs.New(errs.Unauthenticated, err)
			default:
				return ctx, errs.Newf(errs.Unauthenticated, "querybyid: homeID[%s]: %s", homeID, err)
			}
		}

		userID = hme.UserID
		ctx = SetHome(ctx, hme)
	}

	auth := authclient.Authorize{
		Claims: GetClaims(ctx),
		UserID: userID,
		Rule:   auth.RuleAdminOrSubject,
	}

	if err := client.Authorize(ctx, auth); err != nil {
		return ctx, errs.New(errs.Unauthenticated, err)
	}

	return ctx, nil
}
