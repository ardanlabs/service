package mid

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/web/auth"
	"github.com/ardanlabs/service/business/web/errs"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/google/uuid"
)

type ctxUserKey int

const (
	userIDKey ctxUserKey = iota + 1
	userKey
)

// GetUserID returns the claims from the context.
func GetUserID(ctx context.Context) (uuid.UUID, error) {
	v, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}, errors.New("user id not found in context")
	}

	return v, nil
}

// GetUser returns the user from the context.
func GetUser(ctx context.Context) (user.User, error) {
	v, ok := ctx.Value(userKey).(user.User)
	if !ok {
		return user.User{}, errors.New("user not found in context")
	}

	return v, nil
}

func setUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func setUser(ctx context.Context, usr user.User) context.Context {
	return context.WithValue(ctx, userKey, usr)
}

// AuthorizeUser executes the specified role and extracts the specified user
// from the DB if a user id is specified in the call. Depending on the rule
// specified, the userid from the claims may be compared with the specified
// user id.
func AuthorizeUser(a *auth.Auth, userCore *user.Core) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			var userID uuid.UUID

			if id := web.Param(r, "user_id"); id != "" {
				var err error
				userID, err = uuid.Parse(id)
				if err != nil {
					return errs.NewTrusted(ErrInvalidID, http.StatusBadRequest)
				}

				usr, err := userCore.QueryByID(ctx, userID)
				if err != nil {
					switch {
					case errors.Is(err, user.ErrNotFound):
						return errs.NewTrusted(err, http.StatusNoContent)
					default:
						return fmt.Errorf("querybyid: userID[%s]: %w", userID, err)
					}
				}

				ctx = setUser(ctx, usr)
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
