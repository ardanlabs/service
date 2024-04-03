package productgrp

import (
	"net/http"

	"github.com/ardanlabs/service/app/core/crud/productapp"
	"github.com/ardanlabs/service/business/api/page"
)

func parseQueryParams(r *http.Request) (productapp.QueryParams, error) {
	const (
		orderBy          = "orderBy"
		filterPage       = "page"
		filterRow        = "row"
		filterByProdID   = "product_id"
		filterByCost     = "cost"
		filterByQuantity = "quantity"
		filterByName     = "name"
	)

	values := r.URL.Query()

	var filter productapp.QueryParams

	pg, err := page.ParseHTTP(r)
	if err != nil {
		return productapp.QueryParams{}, err
	}
	filter.Page = pg.Number
	filter.Rows = pg.RowsPerPage

	if orderBy := values.Get(orderBy); orderBy != "" {
		filter.OrderBy = orderBy
	}

	if productID := values.Get(filterByProdID); productID != "" {
		filter.ID = productID
	}

	if cost := values.Get(filterByCost); cost != "" {
		filter.Cost = cost
	}

	if quantity := values.Get(filterByQuantity); quantity != "" {
		filter.Quantity = quantity
	}

	if name := values.Get(filterByName); name != "" {
		filter.Name = name
	}

	return filter, nil
}
