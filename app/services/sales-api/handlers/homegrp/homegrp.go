// Package homegrp maintains the group of handlers for home access.
package homegrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/business/core/crud/home"
	"github.com/ardanlabs/service/business/web/errs"
	"github.com/ardanlabs/service/business/web/mid"
	"github.com/ardanlabs/service/business/web/page"
	"github.com/ardanlabs/service/foundation/web"
)

// Set of error variables for handling home group errors.
var (
	ErrInvalidID = errors.New("ID is not in its proper form")
)

type handlers struct {
	home *home.Core
}

func new(home *home.Core) *handlers {
	return &handlers{
		home: home,
	}
}

// create adds a new home to the system.
func (h *handlers) create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app AppNewHome
	if err := web.Decode(r, &app); err != nil {
		return errs.NewTrusted(err, http.StatusBadRequest)
	}

	nh, err := toCoreNewHome(ctx, app)
	if err != nil {
		return errs.NewTrusted(err, http.StatusBadRequest)
	}

	hme, err := h.home.Create(ctx, nh)
	if err != nil {
		return fmt.Errorf("create: hme[%+v]: %w", app, err)
	}

	return web.Respond(ctx, w, toAppHome(hme), http.StatusCreated)
}

// update updates a home in the system.
func (h *handlers) update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app AppUpdateHome
	if err := web.Decode(r, &app); err != nil {
		return errs.NewTrusted(err, http.StatusBadRequest)
	}

	uh, err := toCoreUpdateHome(app)
	if err != nil {
		return errs.NewTrusted(err, http.StatusBadRequest)
	}

	hme, err := mid.GetHome(ctx)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}

	updHme, err := h.home.Update(ctx, hme, uh)
	if err != nil {
		return fmt.Errorf("update: homeID[%s] app[%+v]: %w", hme.ID, app, err)
	}

	return web.Respond(ctx, w, toAppHome(updHme), http.StatusOK)
}

// delete deletes a home from the system.
func (h *handlers) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	hme, err := mid.GetHome(ctx)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	if err := h.home.Delete(ctx, hme); err != nil {
		return fmt.Errorf("delete: homeID[%s]: %w", hme.ID, err)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// query returns a list of homes with paging.
func (h *handlers) query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	pg, err := page.Parse(r)
	if err != nil {
		return err
	}

	filter, err := parseFilter(r)
	if err != nil {
		return err
	}

	orderBy, err := parseOrder(r)
	if err != nil {
		return err
	}

	homes, err := h.home.Query(ctx, filter, orderBy, pg.Number, pg.RowsPerPage)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	total, err := h.home.Count(ctx, filter)
	if err != nil {
		return fmt.Errorf("count: %w", err)
	}

	return web.Respond(ctx, w, page.NewDocument(toAppHomes(homes), total, pg.Number, pg.RowsPerPage), http.StatusOK)
}

// queryByID returns a home by its ID.
func (h *handlers) queryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	hme, err := mid.GetHome(ctx)
	if err != nil {
		return fmt.Errorf("querybyid: %w", err)
	}

	return web.Respond(ctx, w, toAppHome(hme), http.StatusOK)
}
