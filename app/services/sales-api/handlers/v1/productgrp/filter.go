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

	if productID := values.Get("product_id"); productID != "" {
		id, err := uuid.Parse(productID)
		if err != nil {
			return product.QueryFilter{}, validate.NewFieldsError("product_id", err)
		}
		filter.WithProductID(id)
	}

	if cost := values.Get("cost"); cost != "" {
		cst, err := strconv.ParseFloat(cost, 64)
		if err != nil {
			return product.QueryFilter{}, validate.NewFieldsError("cost", err)
		}
		filter.WithCost(cst)
	}

	if quantity := values.Get("quantity"); quantity != "" {
		qua, err := strconv.ParseInt(quantity, 10, 64)
		if err != nil {
			return product.QueryFilter{}, validate.NewFieldsError("quantity", err)
		}
		filter.WithQuantity(int(qua))
	}

	if name := values.Get("name"); name != "" {
		filter.WithName(name)
	}

	if err := filter.Validate(); err != nil {
		return product.QueryFilter{}, err
	}

	return filter, nil
}
