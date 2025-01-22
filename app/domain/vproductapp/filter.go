package vproductapp

import (
	"net/http"
	"strconv"

	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/business/domain/vproductbus"
	"github.com/ardanlabs/service/business/types/name"
	"github.com/google/uuid"
)

type queryParams struct {
	Page     string
	Rows     string
	OrderBy  string
	ID       string
	Name     string
	Cost     string
	Quantity string
	UserName string
}

func parseQueryParams(r *http.Request) queryParams {
	values := r.URL.Query()

	filter := queryParams{
		Page:     values.Get("page"),
		Rows:     values.Get("rows"),
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
	var fieldErrors errs.FieldErrors
	var filter vproductbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		switch err {
		case nil:
			filter.ID = &id
		default:
			fieldErrors.Add("product_id", err)
		}
	}

	if qp.Name != "" {
		name, err := name.Parse(qp.Name)
		switch err {
		case nil:
			filter.Name = &name
		default:
			fieldErrors.Add("name", err)
		}
	}

	if qp.Cost != "" {
		cst, err := strconv.ParseFloat(qp.Cost, 64)
		switch err {
		case nil:
			filter.Cost = &cst
		default:
			fieldErrors.Add("cost", err)
		}
	}

	if qp.Quantity != "" {
		qua, err := strconv.ParseInt(qp.Quantity, 10, 64)
		switch err {
		case nil:
			i := int(qua)
			filter.Quantity = &i
		default:
			fieldErrors.Add("quantity", err)
		}
	}

	if qp.UserName != "" {
		userName, err := name.Parse(qp.UserName)
		switch err {
		case nil:
			filter.UserName = &userName
		default:
			fieldErrors.Add("user_name", err)
		}
	}

	if fieldErrors != nil {
		return vproductbus.QueryFilter{}, fieldErrors.ToError()
	}

	return filter, nil
}
