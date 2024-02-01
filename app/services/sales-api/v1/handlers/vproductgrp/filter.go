package vproductgrp

import (
	"net/http"
	"strconv"

	"github.com/ardanlabs/service/business/core/views/vproduct"
	"github.com/ardanlabs/service/foundation/validate"
	"github.com/google/uuid"
)

func parseFilter(r *http.Request) (vproduct.QueryFilter, error) {
	const (
		filterByProdID   = "product_id"
		filterByCost     = "cost"
		filterByQuantity = "quantity"
		filterByName     = "name"
		filterByUserName = "user_name"
	)

	values := r.URL.Query()

	var filter vproduct.QueryFilter

	if productID := values.Get(filterByProdID); productID != "" {
		id, err := uuid.Parse(productID)
		if err != nil {
			return vproduct.QueryFilter{}, validate.NewFieldsError(filterByProdID, err)
		}
		filter.WithID(id)
	}

	if cost := values.Get(filterByCost); cost != "" {
		cst, err := strconv.ParseFloat(cost, 64)
		if err != nil {
			return vproduct.QueryFilter{}, validate.NewFieldsError(filterByCost, err)
		}
		filter.WithCost(cst)
	}

	if quantity := values.Get(filterByQuantity); quantity != "" {
		qua, err := strconv.ParseInt(quantity, 10, 64)
		if err != nil {
			return vproduct.QueryFilter{}, validate.NewFieldsError(filterByQuantity, err)
		}
		filter.WithQuantity(int(qua))
	}

	if name := values.Get(filterByName); name != "" {
		filter.WithName(name)
	}

	return filter, nil
}
