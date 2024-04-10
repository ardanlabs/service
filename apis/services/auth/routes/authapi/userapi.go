// Package authapi maintains the web based api for auth access.
package authapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/app/core/crud/userapp"
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/foundation/validate"
	"github.com/ardanlabs/service/foundation/web"
)

type api struct {
	userApp *userapp.Core
	authBus *auth.Auth
}

func newAPI(userApp *userapp.Core, authBus *auth.Auth) *api {
	return &api{
		userApp: userApp,
		authBus: authBus,
	}
}

func (api *api) authorize(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var authInfo AuthInfo
	if err := web.Decode(r, &authInfo); err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	if err := api.authBus.Authorize(ctx, authInfo.Claims, authInfo.UserID, authInfo.Rule); err != nil {
		return errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", authInfo.Claims.Roles, authInfo.Rule, err)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (api *api) token(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	kid := web.Param(r, "kid")
	if kid == "" {
		return validate.NewFieldsError("kid", errors.New("missing kid"))
	}

	token, err := api.userApp.Token(ctx, kid)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, token, http.StatusOK)
}
