// Package userapi maintains the web based api for user access.
package userapi

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/app/domain/userapp"
	"github.com/ardanlabs/service/foundation/web"
)

type api struct {
	userApp *userapp.App
}

func newAPI(userApp *userapp.App) *api {
	return &api{
		userApp: userApp,
	}
}

func (api *api) create(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
	var app userapp.NewUser
	if err := web.Decode(r, &app); err != nil {
		return nil, errs.New(errs.FailedPrecondition, err)
	}

	usr, err := api.userApp.Create(ctx, app)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

func (api *api) update(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
	var app userapp.UpdateUser
	if err := web.Decode(r, &app); err != nil {
		return nil, errs.New(errs.FailedPrecondition, err)
	}

	usr, err := api.userApp.Update(ctx, app)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

func (api *api) updateRole(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
	var app userapp.UpdateUserRole
	if err := web.Decode(r, &app); err != nil {
		return nil, errs.New(errs.FailedPrecondition, err)
	}

	usr, err := api.userApp.UpdateRole(ctx, app)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

func (api *api) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
	if err := api.userApp.Delete(ctx); err != nil {
		return nil, err
	}

	return nil, nil
}

func (api *api) query(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
	qp, err := parseQueryParams(r)
	if err != nil {
		return nil, errs.New(errs.FailedPrecondition, err)
	}

	usr, err := api.userApp.Query(ctx, qp)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

func (api *api) queryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
	usr, err := api.userApp.QueryByID(ctx)
	if err != nil {
		return nil, err
	}

	return usr, nil
}
