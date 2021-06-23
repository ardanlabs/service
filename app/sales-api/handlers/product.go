package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ardanlabs/service/business/data/product"
	"github.com/ardanlabs/service/business/sys/auth"
	"github.com/ardanlabs/service/business/sys/validate"
	"github.com/ardanlabs/service/foundation/database"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/pkg/errors"
)

type productGroup struct {
	store product.Store
}

func (pg productGroup) query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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

	products, err := pg.store.Query(ctx, v.TraceID, pageNumber, rowsPerPage)
	if err != nil {
		return errors.Wrap(err, "unable to query for products")
	}

	return web.Respond(ctx, w, products, http.StatusOK)
}

func (pg productGroup) queryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	id := web.Param(r, "id")
	prod, err := pg.store.QueryByID(ctx, v.TraceID, id)
	if err != nil {
		switch errors.Cause(err) {
		case database.ErrInvalidID:
			return validate.NewRequestError(err, http.StatusBadRequest)
		case database.ErrNotFound:
			return validate.NewRequestError(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "ID: %s", id)
		}
	}

	return web.Respond(ctx, w, prod, http.StatusOK)
}

func (pg productGroup) create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return web.NewShutdownError("claims missing from context")
	}

	var np product.NewProduct
	if err := web.Decode(r, &np); err != nil {
		return errors.Wrapf(err, "unable to decode payload")
	}

	prod, err := pg.store.Create(ctx, v.TraceID, claims, np, v.Now)
	if err != nil {
		return errors.Wrapf(err, "creating new product: %+v", np)
	}

	return web.Respond(ctx, w, prod, http.StatusCreated)
}

func (pg productGroup) update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return web.NewShutdownError("claims missing from context")
	}

	var upd product.UpdateProduct
	if err := web.Decode(r, &upd); err != nil {
		return errors.Wrapf(err, "unable to decode payload")
	}

	id := web.Param(r, "id")
	if err := pg.store.Update(ctx, v.TraceID, claims, id, upd, v.Now); err != nil {
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

func (pg productGroup) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	id := web.Param(r, "id")
	if err := pg.store.Delete(ctx, v.TraceID, claims, id); err != nil {
		switch errors.Cause(err) {
		case database.ErrInvalidID:
			return validate.NewRequestError(err, http.StatusBadRequest)
		default:
			return errors.Wrapf(err, "ID: %s", id)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
