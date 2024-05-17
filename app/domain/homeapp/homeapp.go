// Package homeapp maintains the app layer api for the home domain.
package homeapp

import (
	"context"

	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/app/sdk/mid"
	"github.com/ardanlabs/service/app/sdk/query"
	"github.com/ardanlabs/service/business/domain/homebus"
	"github.com/ardanlabs/service/business/sdk/order"
	"github.com/ardanlabs/service/business/sdk/page"
)

// App manages the set of app layer api functions for the home domain.
type App struct {
	homeBus *homebus.Business
}

// NewApp constructs a home domain API for use.
func NewApp(homeBus *homebus.Business) *App {
	return &App{
		homeBus: homeBus,
	}
}

// Create adds a new home to the system.
func (a *App) Create(ctx context.Context, app NewHome) (Home, error) {
	nh, err := toBusNewHome(ctx, app)
	if err != nil {
		return Home{}, errs.New(errs.FailedPrecondition, err)
	}

	hme, err := a.homeBus.Create(ctx, nh)
	if err != nil {
		return Home{}, errs.Newf(errs.Internal, "create: hme[%+v]: %s", app, err)
	}

	return toAppHome(hme), nil
}

// Update updates an existing home.
func (a *App) Update(ctx context.Context, app UpdateHome) (Home, error) {
	uh, err := toBusUpdateHome(app)
	if err != nil {
		return Home{}, errs.New(errs.FailedPrecondition, err)
	}

	hme, err := mid.GetHome(ctx)
	if err != nil {
		return Home{}, errs.Newf(errs.Internal, "home missing in context: %s", err)
	}

	updUsr, err := a.homeBus.Update(ctx, hme, uh)
	if err != nil {
		return Home{}, errs.Newf(errs.Internal, "update: homeID[%s] uh[%+v]: %s", hme.ID, uh, err)
	}

	return toAppHome(updUsr), nil
}

// Delete removes a home from the system.
func (a *App) Delete(ctx context.Context) error {
	hme, err := mid.GetHome(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "homeID missing in context: %s", err)
	}

	if err := a.homeBus.Delete(ctx, hme); err != nil {
		return errs.Newf(errs.Internal, "delete: homeID[%s]: %s", hme.ID, err)
	}

	return nil
}

// Query returns a list of homes with paging.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Home], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Home]{}, err
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Home]{}, err
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Home]{}, err
	}

	hmes, err := a.homeBus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[Home]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.homeBus.Count(ctx, filter)
	if err != nil {
		return query.Result[Home]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toAppHomes(hmes), total, page), nil
}

// QueryByID returns a home by its Ia.
func (a *App) QueryByID(ctx context.Context) (Home, error) {
	hme, err := mid.GetHome(ctx)
	if err != nil {
		return Home{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return toAppHome(hme), nil
}
