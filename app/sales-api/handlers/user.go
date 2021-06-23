package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ardanlabs/service/business/data/user"
	"github.com/ardanlabs/service/business/sys/auth"
	"github.com/ardanlabs/service/business/sys/validate"
	"github.com/ardanlabs/service/foundation/database"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/pkg/errors"
)

type userGroup struct {
	store user.Store
	auth  *auth.Auth
}

func (ug userGroup) query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	page := web.Param(r, "page")
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		return validate.NewRequestError(fmt.Errorf("invalid page format: %s", page), http.StatusBadRequest)
	}
	rows := web.Param(r, "rows")
	rowsPerPage, err := strconv.Atoi(rows)
	if err != nil {
		return validate.NewRequestError(fmt.Errorf("invalid rows format: %s", rows), http.StatusBadRequest)
	}

	users, err := ug.store.Query(ctx, v.TraceID, pageNumber, rowsPerPage)
	if err != nil {
		return errors.Wrap(err, "unable to query for users")
	}

	return web.Respond(ctx, w, users, http.StatusOK)
}

func (ug userGroup) queryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	id := web.Param(r, "id")
	usr, err := ug.store.QueryByID(ctx, v.TraceID, claims, id)
	if err != nil {
		switch errors.Cause(err) {
		case database.ErrInvalidID:
			return validate.NewRequestError(err, http.StatusBadRequest)
		case database.ErrNotFound:
			return validate.NewRequestError(err, http.StatusNotFound)
		case database.ErrForbidden:
			return validate.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "ID: %s", id)
		}
	}

	return web.Respond(ctx, w, usr, http.StatusOK)
}

func (ug userGroup) create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	var nu user.NewUser
	if err := web.Decode(r, &nu); err != nil {
		return errors.Wrap(err, "unable to decode payload")
	}

	usr, err := ug.store.Create(ctx, v.TraceID, nu, v.Now)
	if err != nil {
		return errors.Wrapf(err, "User: %+v", &usr)
	}

	return web.Respond(ctx, w, usr, http.StatusCreated)
}

func (ug userGroup) update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	var upd user.UpdateUser
	if err := web.Decode(r, &upd); err != nil {
		return errors.Wrap(err, "unable to decode payload")
	}

	id := web.Param(r, "id")
	err := ug.store.Update(ctx, v.TraceID, claims, id, upd, v.Now)
	if err != nil {
		switch errors.Cause(err) {
		case database.ErrInvalidID:
			return validate.NewRequestError(err, http.StatusBadRequest)
		case database.ErrNotFound:
			return validate.NewRequestError(err, http.StatusNotFound)
		case database.ErrForbidden:
			return validate.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "ID: %s  User: %+v", id, &upd)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (ug userGroup) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	id := web.Param(r, "id")
	err := ug.store.Delete(ctx, v.TraceID, claims, id)
	if err != nil {
		switch errors.Cause(err) {
		case database.ErrInvalidID:
			return validate.NewRequestError(err, http.StatusBadRequest)
		case database.ErrNotFound:
			return validate.NewRequestError(err, http.StatusNotFound)
		case database.ErrForbidden:
			return validate.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "ID: %s", id)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (ug userGroup) token(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	email, pass, ok := r.BasicAuth()
	if !ok {
		err := errors.New("must provide email and password in Basic auth")
		return validate.NewRequestError(err, http.StatusUnauthorized)
	}

	claims, err := ug.store.Authenticate(ctx, v.TraceID, v.Now, email, pass)
	if err != nil {
		switch errors.Cause(err) {
		case database.ErrAuthenticationFailure:
			return validate.NewRequestError(err, http.StatusUnauthorized)
		default:
			return errors.Wrap(err, "authenticating")
		}
	}

	kid := web.Param(r, "kid")
	var tkn struct {
		Token string `json:"token"`
	}
	tkn.Token, err = ug.auth.GenerateToken(kid, claims)
	if err != nil {
		return errors.Wrap(err, "generating token")
	}

	return web.Respond(ctx, w, tkn, http.StatusOK)
}
