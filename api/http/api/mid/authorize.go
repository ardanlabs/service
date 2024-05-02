package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/api/http/api/response"
	"github.com/ardanlabs/service/app/api/authclient"
	"github.com/ardanlabs/service/app/api/mid"
	"github.com/ardanlabs/service/business/domain/homebus"
	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// Authorize validates authorization via the auth service.
func Authorize(log *logger.Logger, client *authclient.Client, rule string) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
			hdl := func(ctx context.Context) mid.Response {
				return response.ToMid(handler(ctx, w, r))
			}

			return response.ToWebX("Authorize", mid.Authorize(ctx, log, client, rule, hdl))
		}

		return h
	}

	return m
}

// AuthorizeUser executes the specified role and extracts the specified
// user from the DB if a user id is specified in the call. Depending on the rule
// specified, the userid from the claims may be compared with the specified
// user id.
func AuthorizeUser(log *logger.Logger, client *authclient.Client, userBus *userbus.Business, rule string) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
			hdl := func(ctx context.Context) mid.Response {
				return response.ToMid(handler(ctx, w, r))
			}

			return response.ToWeb(mid.AuthorizeUser(ctx, log, client, userBus, rule, web.Param(r, "user_id"), hdl))
		}

		return h
	}

	return m
}

// AuthorizeProduct executes the specified role and extracts the specified
// product from the DB if a product id is specified in the call. Depending on
// the rule specified, the userid from the claims may be compared with the
// specified user id from the product.
func AuthorizeProduct(log *logger.Logger, client *authclient.Client, productBus *productbus.Business) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
			hdl := func(ctx context.Context) mid.Response {
				return response.ToMid(handler(ctx, w, r))
			}

			return response.ToWeb(mid.AuthorizeProduct(ctx, log, client, productBus, web.Param(r, "product_id"), hdl))
		}

		return h
	}

	return m
}

// AuthorizeHome executes the specified role and extracts the specified
// home from the DB if a home id is specified in the call. Depending on
// the rule specified, the userid from the claims may be compared with the
// specified user id from the home.
func AuthorizeHome(log *logger.Logger, client *authclient.Client, homeBus *homebus.Business) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
			hdl := func(ctx context.Context) mid.Response {
				return response.ToMid(handler(ctx, w, r))
			}

			return response.ToWeb(mid.AuthorizeHome(ctx, log, client, homeBus, web.Param(r, "home_id"), hdl))
		}

		return h
	}

	return m
}
