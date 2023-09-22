// Package homegrp maintains the group of handlers for home access.
package homegrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/paging"
	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/request"
	"github.com/ardanlabs/service/business/core/home"
	"github.com/ardanlabs/service/business/data/transaction"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/google/uuid"
)

// Set of error variables for handling home group errors.
var (
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// Handlers manages the set of home enpoints.
type Handlers struct {
	home *home.Core
}

// New constructs a handlers for route access.
func New(home *home.Core) *Handlers {
	return &Handlers{
		home: home,
	}
}

// executeUnderTransaction constructs a new Handlers value with the core apis
// using a store transaction that was created via middleware.
func (h *Handlers) executeUnderTransaction(ctx context.Context) (*Handlers, error) {
	if tx, ok := transaction.Get(ctx); ok {
		home, err := h.home.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		h = &Handlers{
			home: home,
		}

		return h, nil
	}

	return h, nil
}

// Query returns a list of homes with paging.
func (h *Handlers) Query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page, err := paging.ParseRequest(r)
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

	homes, err := h.home.Query(ctx, filter, orderBy, page.Number, page.RowsPerPage)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	total, err := h.home.Count(ctx, filter)
	if err != nil {
		return fmt.Errorf("count: %w", err)
	}

	return web.Respond(ctx, w, paging.NewResponse(toAppHomes(homes), total, page.Number, page.RowsPerPage), http.StatusOK)
}

// QueryByID returns a home by its ID.
func (h *Handlers) QueryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	homeID, err := uuid.Parse(web.Param(r, "home_id"))
	if err != nil {
		return request.NewError(ErrInvalidID, http.StatusBadRequest)
	}

	hme, err := h.home.QueryByID(ctx, homeID)
	if err != nil {
		switch {
		case errors.Is(err, home.ErrNotFound):
			return request.NewError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("querybyid: homeID[%s] %w", homeID, err)
		}
	}

	return web.Respond(ctx, w, toAppHome(hme), http.StatusOK)
}

// Create adds a new home to the system.
func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app AppNewHome
	if err := web.Decode(r, &app); err != nil {
		return request.NewError(err, http.StatusBadRequest)
	}

	nh, err := toCoreNewHome(app)
	if err != nil {
		return request.NewError(err, http.StatusBadRequest)
	}

	hme, err := h.home.Create(ctx, nh)
	if err != nil {
		return fmt.Errorf("create: hme[%+v]: %w", app, err)
	}

	return web.Respond(ctx, w, toAppHome(hme), http.StatusCreated)
}

// Update updates a home in the system.
func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return err
	}

	var app AppUpdateHome
	if err := web.Decode(r, &app); err != nil {
		return request.NewError(err, http.StatusBadRequest)
	}

	homeID, err := uuid.Parse(web.Param(r, "home_id"))
	if err != nil {
		return request.NewError(ErrInvalidID, http.StatusBadRequest)
	}

	hme, err := h.home.QueryByID(ctx, homeID)
	if err != nil {
		switch {
		case errors.Is(err, home.ErrNotFound):
			return request.NewError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("querybyid: homeID[%s] %w", homeID, err)
		}
	}

	hme, err = h.home.Update(ctx, hme, toCoreUpdateHome(app))
	if err != nil {
		return fmt.Errorf("update: homeID[%s] app[%+v]: %w", homeID, app, err)
	}

	return web.Respond(ctx, w, toAppHome(hme), http.StatusOK)
}

// Delete deletes a home from the system.
func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h, err := h.executeUnderTransaction(ctx)
	if err != nil {
		return err
	}

	homeID, err := uuid.Parse(web.Param(r, "home_id"))
	if err != nil {
		return request.NewError(ErrInvalidID, http.StatusBadRequest)
	}

	hme, err := h.home.QueryByID(ctx, homeID)
	if err != nil {
		switch {
		case errors.Is(err, home.ErrNotFound):
			return request.NewError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("querybyid: homeID[%s] %w", homeID, err)
		}
	}

	if err = h.home.Delete(ctx, hme); err != nil {
		return fmt.Errorf("delete: homeID[%s]: %w", homeID, err)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
