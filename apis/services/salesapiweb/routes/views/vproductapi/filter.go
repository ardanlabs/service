package vproductapi

import (
	"net/http"

	"github.com/ardanlabs/service/app/core/views/vproductapp"
	"github.com/ardanlabs/service/business/api/page"
)

func parseQueryParams(r *http.Request) (vproductapp.QueryParams, error) {
	const (
		orderBy          = "orderBy"
		filterPage       = "page"
		filterRow        = "row"
		filterByProdID   = "product_id"
		filterByCost     = "cost"
		filterByQuantity = "quantity"
		filterByName     = "name"
		filterByUserName = "user_name"
	)

	values := r.URL.Query()

	var filter vproductapp.QueryParams

	pg, err := page.ParseHTTP(r)
	if err != nil {
		return vproductapp.QueryParams{}, err
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

	if userName := values.Get(filterByUserName); userName != "" {
		filter.UserName = userName
	}

	return filter, nil
}
