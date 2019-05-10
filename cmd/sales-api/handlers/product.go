package handlers

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/internal/platform/db"
	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/ardanlabs/service/internal/product"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

// Product represents the Product API method handler set.
type Product struct {
	MasterDB *db.DB

	// ADD OTHER STATE LIKE THE LOGGER IF NEEDED.
}

// List returns all the existing products in the system.
func (p *Product) List(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Product.List")
	defer span.End()

	dbConn := p.MasterDB.Copy()
	defer dbConn.Close()

	products, err := product.List(ctx, dbConn)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, products, http.StatusOK)
}

// Retrieve returns the specified product from the system.
func (p *Product) Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Product.Retrieve")
	defer span.End()

	dbConn := p.MasterDB.Copy()
	defer dbConn.Close()

	prod, err := product.Retrieve(ctx, dbConn, params["id"])
	if err != nil {
		switch err {
		case product.ErrInvalidID:
			return web.ErrRequestFailed(err, http.StatusBadRequest)
		case product.ErrNotFound:
			return web.ErrRequestFailed(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "ID: %s", params["id"])
		}
	}

	return web.Respond(ctx, w, prod, http.StatusOK)
}

// Create inserts a new product into the system.
func (p *Product) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Product.Create")
	defer span.End()

	dbConn := p.MasterDB.Copy()
	defer dbConn.Close()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.ErrShutdown("web value missing from context")
	}

	var np product.NewProduct
	if err := web.Decode(r, &np); err != nil {
		return errors.Wrap(err, "")
	}

	nUsr, err := product.Create(ctx, dbConn, &np, v.Now)
	if err != nil {
		return errors.Wrapf(err, "Product: %+v", &np)
	}

	return web.Respond(ctx, w, nUsr, http.StatusCreated)
}

// Update updates the specified product in the system.
func (p *Product) Update(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Product.Update")
	defer span.End()

	dbConn := p.MasterDB.Copy()
	defer dbConn.Close()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.ErrShutdown("web value missing from context")
	}

	var up product.UpdateProduct
	if err := web.Decode(r, &up); err != nil {
		return errors.Wrap(err, "")
	}

	err := product.Update(ctx, dbConn, params["id"], up, v.Now)
	if err != nil {
		switch err {
		case product.ErrInvalidID:
			return web.ErrRequestFailed(err, http.StatusBadRequest)
		case product.ErrNotFound:
			return web.ErrRequestFailed(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "ID: %s Update: %+v", params["id"], up)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Delete removes the specified product from the system.
func (p *Product) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Product.Delete")
	defer span.End()

	dbConn := p.MasterDB.Copy()
	defer dbConn.Close()

	err := product.Delete(ctx, dbConn, params["id"])
	if err != nil {
		switch err {
		case product.ErrInvalidID:
			return web.ErrRequestFailed(err, http.StatusBadRequest)
		case product.ErrNotFound:
			return web.ErrRequestFailed(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "Id: %s", params["id"])
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
