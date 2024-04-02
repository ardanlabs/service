package usergrp

import (
	"net/http"

	"github.com/ardanlabs/service/app/services/sales-api/apis/crud/userapi"
	"github.com/ardanlabs/service/business/api/page"
)

func parseQueryParams(r *http.Request) (userapi.QueryParams, error) {
	const (
		orderBy                  = "orderBy"
		filterPage               = "page"
		filterRow                = "row"
		filterByUserID           = "user_id"
		filterByEmail            = "email"
		filterByStartCreatedDate = "start_created_date"
		filterByEndCreatedDate   = "end_created_date"
		filterByName             = "name"
	)

	values := r.URL.Query()

	var filter userapi.QueryParams

	pg, err := page.Parse(r)
	if err != nil {
		return userapi.QueryParams{}, err
	}
	filter.Page = pg.Number
	filter.Rows = pg.RowsPerPage

	if orderBy := values.Get(orderBy); orderBy != "" {
		filter.OrderBy = orderBy
	}

	if userID := values.Get(filterByUserID); userID != "" {
		filter.ID = userID
	}

	if email := values.Get(filterByEmail); email != "" {
		filter.Email = email
	}

	if startedDate := values.Get(filterByStartCreatedDate); startedDate != "" {
		filter.StartCreatedDate = startedDate
	}

	if endDate := values.Get(filterByStartCreatedDate); endDate != "" {
		filter.EndCreatedDate = endDate
	}

	return filter, nil
}
