// Package authapi maintains the web based api for auth access.
package authapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/ardanlabs/service/app/api/auth"
	"github.com/ardanlabs/service/app/api/authclient"
	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/app/api/mid"
	"github.com/ardanlabs/service/foundation/validate"
	"github.com/ardanlabs/service/foundation/web"
)

type api struct {
	auth *auth.Auth
}

func newAPI(ath *auth.Auth) *api {
	return &api{
		auth: ath,
	}
}

func (api *api) token(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
	kid := web.Param(r, "kid")
	if kid == "" {
		return nil, errs.New(errs.FailedPrecondition, validate.NewFieldsError("kid", errors.New("missing kid")))
	}

	// The BearerBasic middleware function generates the claims.
	claims := mid.GetClaims(ctx)

	tkn, err := api.auth.GenerateToken(kid, claims)
	if err != nil {
		return nil, errs.New(errs.Internal, err)
	}

	token := struct {
		Token string `json:"token"`
	}{
		Token: tkn,
	}

	return token, nil
}

func (api *api) authenticate(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
	// The middleware is actually handling the authentication. So if the code
	// gets to this handler, authentication passed.

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return nil, errs.New(errs.Unauthenticated, err)
	}

	resp := authclient.AuthenticateResp{
		UserID: userID,
		Claims: mid.GetClaims(ctx),
	}

	return resp, nil
}

func (api *api) authorize(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
	var auth authclient.Authorize
	if err := web.Decode(r, &auth); err != nil {
		return nil, errs.New(errs.FailedPrecondition, err)
	}

	if err := api.auth.Authorize(ctx, auth.Claims, auth.UserID, auth.Rule); err != nil {
		return nil, errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", auth.Claims.Roles, auth.Rule, err)
	}

	return nil, nil
}
