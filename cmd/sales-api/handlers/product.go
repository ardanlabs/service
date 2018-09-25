package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/ardanlabs/service/internal/platform/db"
	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/ardanlabs/service/internal/product"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

// Product represents the Product API method handler set.
type Product struct {
	MasterDB *db.DB

	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// List returns all the existing products in the system.
func (p *Product) List(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Product.List")
	defer span.End()

	dbConn := p.MasterDB.Copy()
	defer dbConn.Close()

	products, err := product.List(ctx, dbConn)
	if err = translate(err); err != nil {
		return errors.Wrap(err, "")
	}

	web.Respond(ctx, log, w, products, http.StatusOK)
	return nil
}

// Retrieve returns the specified product from the system.
func (p *Product) Retrieve(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Product.Retrieve")
	defer span.End()

	dbConn := p.MasterDB.Copy()
	defer dbConn.Close()

	prod, err := product.Retrieve(ctx, dbConn, params["id"])
	if err = translate(err); err != nil {
		return errors.Wrapf(err, "ID: %s", params["id"])
	}

	web.Respond(ctx, log, w, prod, http.StatusOK)
	return nil
}

// Create inserts a new product into the system.
func (p *Product) Create(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Product.Create")
	defer span.End()

	dbConn := p.MasterDB.Copy()
	defer dbConn.Close()

	var np product.NewProduct
	if err := web.Unmarshal(r.Body, &np); err != nil {
		return errors.Wrap(err, "")
	}

	// TODO(jlw) use time from request context
	nUsr, err := product.Create(ctx, dbConn, &np, time.Now().UTC())
	if err = translate(err); err != nil {
		return errors.Wrapf(err, "Product: %+v", &np)
	}

	web.Respond(ctx, log, w, nUsr, http.StatusCreated)
	return nil
}

// Update updates the specified product in the system.
func (p *Product) Update(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Product.Update")
	defer span.End()

	dbConn := p.MasterDB.Copy()
	defer dbConn.Close()

	var up product.UpdateProduct
	if err := web.Unmarshal(r.Body, &up); err != nil {
		return errors.Wrap(err, "")
	}

	// TODO(jlw) use time from request context
	err := product.Update(ctx, dbConn, params["id"], up, time.Now().UTC())
	if err = translate(err); err != nil {
		return errors.Wrapf(err, "ID: %s Update: %+v", params["id"], up)
	}

	web.Respond(ctx, log, w, nil, http.StatusNoContent)
	return nil
}

// Delete removed the specified product from the system.
func (p *Product) Delete(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Product.Delete")
	defer span.End()

	dbConn := p.MasterDB.Copy()
	defer dbConn.Close()

	err := product.Delete(ctx, dbConn, params["id"])
	if err = translate(err); err != nil {
		return errors.Wrapf(err, "Id: %s", params["id"])
	}

	web.Respond(ctx, log, w, nil, http.StatusNoContent)
	return nil
}
