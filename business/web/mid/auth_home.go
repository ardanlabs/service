package mid

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/business/core/crud/home"
	"github.com/ardanlabs/service/business/web/auth"
	"github.com/ardanlabs/service/business/web/errs"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/google/uuid"
)

type ctxHomeKey int

const homeKey ctxHomeKey = 1

// GetHome returns the home from the context.
func GetHome(ctx context.Context) (home.Home, error) {
	v, ok := ctx.Value(homeKey).(home.Home)
	if !ok {
		return home.Home{}, errors.New("home not found in context")
	}

	return v, nil
}

func setHome(ctx context.Context, hme home.Home) context.Context {
	return context.WithValue(ctx, homeKey, hme)
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
					return errs.NewTrusted(ErrInvalidID, http.StatusBadRequest)
				}

				hme, err := homeCore.QueryByID(ctx, homeID)
				if err != nil {
					switch {
					case errors.Is(err, home.ErrNotFound):
						return errs.NewTrusted(err, http.StatusNoContent)
					default:
						return fmt.Errorf("querybyid: homeID[%s]: %w", homeID, err)
					}
				}

				userID = hme.UserID
				ctx = setHome(ctx, hme)
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
