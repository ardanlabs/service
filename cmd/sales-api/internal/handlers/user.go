package handlers

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/internal/data"
	"github.com/ardanlabs/service/internal/platform/auth"
	"github.com/ardanlabs/service/internal/platform/trace"
	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type user struct {
	db            *sqlx.DB
	authenticator *auth.Authenticator
}

func (u *user) list(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx = trace.NewSpan(ctx, "handlers.user.list")

	users, err := data.Retrieve.User.List(ctx, u.db)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, users, http.StatusOK)
}

func (u *user) retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx = trace.NewSpan(ctx, "handlers.user.retrieve")

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	usr, err := data.Retrieve.User.One(ctx, claims, u.db, params["id"])
	if err != nil {
		switch err {
		case data.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case data.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case data.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "Id: %s", params["id"])
		}
	}

	return web.Respond(ctx, w, usr, http.StatusOK)
}

func (u *user) create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx = trace.NewSpan(ctx, "handlers.user.create")

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	var nu data.NewUser
	if err := web.Decode(r, &nu); err != nil {
		return errors.Wrap(err, "")
	}

	usr, err := data.Create.User(ctx, u.db, nu, v.Now)
	if err != nil {
		return errors.Wrapf(err, "User: %+v", &usr)
	}

	return web.Respond(ctx, w, usr, http.StatusCreated)
}

func (u *user) update(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx = trace.NewSpan(ctx, "handlers.user.update")

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	var upd data.UpdateUser
	if err := web.Decode(r, &upd); err != nil {
		return errors.Wrap(err, "")
	}

	err := data.Update.User(ctx, claims, u.db, params["id"], upd, v.Now)
	if err != nil {
		switch err {
		case data.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case data.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case data.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "ID: %s  User: %+v", params["id"], &upd)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (u *user) delete(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx = trace.NewSpan(ctx, "handlers.user.delete")

	err := data.Delete.User(ctx, u.db, params["id"])
	if err != nil {
		switch err {
		case data.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case data.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case data.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "Id: %s", params["id"])
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (u *user) token(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx = trace.NewSpan(ctx, "handlers.User.Token")

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	email, pass, ok := r.BasicAuth()
	if !ok {
		err := errors.New("must provide email and password in Basic auth")
		return web.NewRequestError(err, http.StatusUnauthorized)
	}

	claims, err := data.Authenticate(ctx, u.db, v.Now, email, pass)
	if err != nil {
		switch err {
		case data.ErrAuthenticationFailure:
			return web.NewRequestError(err, http.StatusUnauthorized)
		default:
			return errors.Wrap(err, "authenticating")
		}
	}

	var tkn struct {
		Token string `json:"token"`
	}
	tkn.Token, err = u.authenticator.GenerateToken(claims)
	if err != nil {
		return errors.Wrap(err, "generating token")
	}

	return web.Respond(ctx, w, tkn, http.StatusOK)
}
