package handlers

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/internal/data"
	"github.com/ardanlabs/service/internal/platform/auth"
	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/dimfeld/httptreemux/v5"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/api/global"
)

type product struct {
	db *sqlx.DB
}

func (p *product) list(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := global.Tracer("service").Start(ctx, "handlers.product.list")
	defer span.End()

	products, err := data.Retrieve.Product.List(ctx, p.db)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, products, http.StatusOK)
}

func (p *product) retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := global.Tracer("service").Start(ctx, "handlers.product.retrieve")
	defer span.End()

	params := httptreemux.ContextParams(r.Context())
	prod, err := data.Retrieve.Product.One(ctx, p.db, params["id"])
	if err != nil {
		switch err {
		case data.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case data.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "ID: %s", params["id"])
		}
	}

	return web.Respond(ctx, w, prod, http.StatusOK)
}

func (p *product) create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := global.Tracer("service").Start(ctx, "handlers.product.create")
	defer span.End()

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return web.NewShutdownError("claims missing from context")
	}

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	var np data.NewProduct
	if err := web.Decode(r, &np); err != nil {
		return errors.Wrap(err, "decoding new product")
	}

	prod, err := data.Create.Product(ctx, p.db, claims, np, v.Now)
	if err != nil {
		return errors.Wrapf(err, "creating new product: %+v", np)
	}

	return web.Respond(ctx, w, prod, http.StatusCreated)
}

func (p *product) update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := global.Tracer("service").Start(ctx, "handlers.product.update")
	defer span.End()

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return web.NewShutdownError("claims missing from context")
	}

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	var up data.UpdateProduct
	if err := web.Decode(r, &up); err != nil {
		return errors.Wrap(err, "")
	}

	params := httptreemux.ContextParams(r.Context())
	if err := data.Update.Product(ctx, p.db, claims, params["id"], up, v.Now); err != nil {
		switch err {
		case data.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case data.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case data.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "updating product %q: %+v", params["id"], up)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (p *product) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := global.Tracer("service").Start(ctx, "handlers.product.delete")
	defer span.End()

	params := httptreemux.ContextParams(r.Context())
	if err := data.Delete.Product(ctx, p.db, params["id"]); err != nil {
		switch err {
		case data.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		default:
			return errors.Wrapf(err, "Id: %s", params["id"])
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
