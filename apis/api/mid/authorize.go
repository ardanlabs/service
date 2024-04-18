package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/apis/api/authclient"
	"github.com/ardanlabs/service/app/api/mid"
	"github.com/ardanlabs/service/business/domain/homebus"
	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// Authorize executes the authorize middleware functionality.
func Authorize(log *logger.Logger, client *authclient.Client, rule string) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			if err := mid.Authorize(ctx, log, client, rule); err != nil {
				return err
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

// AuthorizeUser executes the authorize user middleware functionality.
func AuthorizeUser(log *logger.Logger, client *authclient.Client, userBus *userbus.Core, rule string) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx, err := mid.AuthorizeUser(ctx, log, client, userBus, rule, web.Param(r, "user_id"))
			if err != nil {
				return err
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

// AuthorizeProduct executes the authorize product middleware functionality.
func AuthorizeProduct(log *logger.Logger, client *authclient.Client, productBus *productbus.Core) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx, err := mid.AuthorizeProduct(ctx, log, client, productBus, web.Param(r, "product_id"))
			if err != nil {
				return err
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

// AuthorizeHome executes the authorize home middleware functionality.
func AuthorizeHome(log *logger.Logger, client *authclient.Client, homeBus *homebus.Core) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx, err := mid.AuthorizeHome(ctx, log, client, homeBus, web.Param(r, "home_id"))
			if err != nil {
				return err
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
