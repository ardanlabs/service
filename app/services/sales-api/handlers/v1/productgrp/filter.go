package productgrp

import (
	"net/http"
	"strconv"

	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/sys/validate"
	"github.com/google/uuid"
)

func parseFilter(r *http.Request) (product.QueryFilter, error) {
	values := r.URL.Query()

	var filter product.QueryFilter

	if id, err := uuid.Parse(values.Get("id")); err == nil {
		filter.ByID(id)
	}

	if err := filter.ByName(values.Get("name")); err != nil {
		return product.QueryFilter{}, validate.NewFieldsError("name", err)
	}

	cost := values.Get("cost")
	if cost != "" {
		cst, err := strconv.ParseInt(cost, 10, 64)
		if err != nil {
			return product.QueryFilter{}, validate.NewFieldsError("cost", err)
		}

		filter.ByCost(int(cst))
	}

	quantity := values.Get("quantity")
	if quantity != "" {
		qua, err := strconv.ParseInt(quantity, 10, 64)
		if err != nil {
			return product.QueryFilter{}, validate.NewFieldsError("quantity", err)
		}

		filter.ByQuantity(int(qua))
	}

	if err := filter.Validate(); err != nil {
		return product.QueryFilter{}, err
	}

	return filter, nil
}
