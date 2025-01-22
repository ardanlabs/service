package userapp

import (
	"net/http"
	"net/mail"
	"time"

	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/business/types/name"
	"github.com/google/uuid"
)

type queryParams struct {
	Page             string
	Rows             string
	OrderBy          string
	ID               string
	Name             string
	Email            string
	StartCreatedDate string
	EndCreatedDate   string
}

func parseQueryParams(r *http.Request) (queryParams, error) {
	values := r.URL.Query()

	filter := queryParams{
		Page:             values.Get("page"),
		Rows:             values.Get("rows"),
		OrderBy:          values.Get("orderBy"),
		ID:               values.Get("user_id"),
		Name:             values.Get("name"),
		Email:            values.Get("email"),
		StartCreatedDate: values.Get("start_created_date"),
		EndCreatedDate:   values.Get("end_created_date"),
	}

	return filter, nil
}

func parseFilter(qp queryParams) (userbus.QueryFilter, error) {
	var fieldErrors errs.FieldErrors
	var filter userbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		switch err {
		case nil:
			filter.ID = &id
		default:
			fieldErrors.Add("user_id", err)
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

	if qp.Email != "" {
		addr, err := mail.ParseAddress(qp.Email)
		switch err {
		case nil:
			filter.Email = addr
		default:
			fieldErrors.Add("email", err)
		}
	}

	if qp.StartCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.StartCreatedDate)
		switch err {
		case nil:
			filter.StartCreatedDate = &t
		default:
			fieldErrors.Add("start_created_date", err)
		}
	}

	if qp.EndCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.EndCreatedDate)
		switch err {
		case nil:
			filter.EndCreatedDate = &t
		default:
			fieldErrors.Add("end_created_date", err)
		}
	}

	if fieldErrors != nil {
		return userbus.QueryFilter{}, fieldErrors.ToError()
	}

	return filter, nil
}
