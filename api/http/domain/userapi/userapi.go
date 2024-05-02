// Package userapi maintains the web based api for user access.
package userapi

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/api/http/api/response"
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

func (api *api) create(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
	var app userapp.NewUser
	if err := web.Decode(r, &app); err != nil {
		return response.AppError(errs.FailedPrecondition, err)
	}

	usr, err := api.userApp.Create(ctx, app)
	if err != nil {
		return response.AppError(errs.Internal, err)
	}

	return response.Response(usr, http.StatusCreated)
}

func (api *api) update(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
	var app userapp.UpdateUser
	if err := web.Decode(r, &app); err != nil {
		return response.AppError(errs.FailedPrecondition, err)
	}

	usr, err := api.userApp.Update(ctx, app)
	if err != nil {
		return response.AppError(errs.Internal, err)
	}

	return response.Response(usr, http.StatusOK)
}

func (api *api) updateRole(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
	var app userapp.UpdateUserRole
	if err := web.Decode(r, &app); err != nil {
		return response.AppError(errs.FailedPrecondition, err)
	}

	usr, err := api.userApp.UpdateRole(ctx, app)
	if err != nil {
		return response.AppError(errs.Internal, err)
	}

	return response.Response(usr, http.StatusOK)
}

func (api *api) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
	if err := api.userApp.Delete(ctx); err != nil {
		return response.AppError(errs.Internal, err)
	}

	return response.Response(nil, http.StatusNoContent)
}

func (api *api) query(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
	qp, err := parseQueryParams(r)
	if err != nil {
		return response.AppError(errs.Internal, err)
	}

	usr, err := api.userApp.Query(ctx, qp)
	if err != nil {
		return response.AppError(errs.Internal, err)
	}

	return response.Response(usr, http.StatusOK)
}

func (api *api) queryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Response {
	usr, err := api.userApp.QueryByID(ctx)
	if err != nil {
		return response.AppError(errs.Internal, err)
	}

	return response.Response(usr, http.StatusOK)
}
