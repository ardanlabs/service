package productgrp

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ardanlabs/service/business/core/product"
)

func getFilter(r *http.Request) (product.QueryFilter, error) {
	values := r.URL.Query()

	var filter product.QueryFilter
	filter.ByID(values.Get("id"))
	filter.ByName(values.Get("name"))

	cost := values.Get("cost")
	if cost != "" {
		cst, err := strconv.ParseInt(cost, 10, 64)
		if err != nil {
			return product.QueryFilter{}, fmt.Errorf("invalid field filter cost format: %s", cost)
		}

		filter.ByCost(int(cst))
	}

	quantity := values.Get("quantity")
	if quantity != "" {
		qua, err := strconv.ParseInt(quantity, 10, 64)
		if err != nil {
			return product.QueryFilter{}, fmt.Errorf("invalid field filter quantity format: %s", quantity)
		}

		filter.ByQuantity(int(qua))
	}

	return filter, nil
}
