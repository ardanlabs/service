package vproductapp

import (
	"net/http"
	"strconv"

	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/business/domain/vproductbus"
	"github.com/ardanlabs/service/business/types/name"
	"github.com/google/uuid"
)

func parseQueryParams(r *http.Request) queryParams {
	values := r.URL.Query()

	filter := queryParams{
		Page:     values.Get("page"),
		Rows:     values.Get("row"),
		OrderBy:  values.Get("orderBy"),
		ID:       values.Get("product_id"),
		Name:     values.Get("name"),
		Cost:     values.Get("cost"),
		Quantity: values.Get("quantity"),
		UserName: values.Get("user_name"),
	}

	return filter
}

func parseFilter(qp queryParams) (vproductbus.QueryFilter, error) {
	var filter vproductbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return vproductbus.QueryFilter{}, errs.NewFieldsError("product_id", err)
		}
		filter.ID = &id
	}

	if qp.Name != "" {
		name, err := name.Parse(qp.Name)
		if err != nil {
			return vproductbus.QueryFilter{}, errs.NewFieldsError("name", err)
		}
		filter.Name = &name
	}

	if qp.Cost != "" {
		cst, err := strconv.ParseFloat(qp.Cost, 64)
		if err != nil {
			return vproductbus.QueryFilter{}, errs.NewFieldsError("cost", err)
		}
		filter.Cost = &cst
	}

	if qp.Quantity != "" {
		qua, err := strconv.ParseInt(qp.Quantity, 10, 64)
		if err != nil {
			return vproductbus.QueryFilter{}, errs.NewFieldsError("quantity", err)
		}
		i := int(qua)
		filter.Quantity = &i
	}

	if qp.Name != "" {
		name, err := name.Parse(qp.Name)
		if err != nil {
			return vproductbus.QueryFilter{}, errs.NewFieldsError("name", err)
		}
		filter.UserName = &name
	}

	return filter, nil
}
