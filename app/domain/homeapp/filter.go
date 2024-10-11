package homeapp

import (
	"net/http"
	"time"

	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/business/domain/homebus"
	"github.com/ardanlabs/service/business/types/hometype"
	"github.com/google/uuid"
)

func parseQueryParams(r *http.Request) queryParams {
	values := r.URL.Query()

	filter := queryParams{
		Page:             values.Get("page"),
		Rows:             values.Get("row"),
		OrderBy:          values.Get("orderBy"),
		ID:               values.Get("home_id"),
		UserID:           values.Get("user_id"),
		Type:             values.Get("type"),
		StartCreatedDate: values.Get("start_created_date"),
		EndCreatedDate:   values.Get("end_created_date"),
	}

	return filter
}

func parseFilter(qp queryParams) (homebus.QueryFilter, error) {
	var filter homebus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return homebus.QueryFilter{}, errs.NewFieldsError("home_id", err)
		}
		filter.ID = &id
	}

	if qp.UserID != "" {
		id, err := uuid.Parse(qp.UserID)
		if err != nil {
			return homebus.QueryFilter{}, errs.NewFieldsError("user_id", err)
		}
		filter.UserID = &id
	}

	if qp.Type != "" {
		typ, err := hometype.Parse(qp.Type)
		if err != nil {
			return homebus.QueryFilter{}, errs.NewFieldsError("type", err)
		}
		filter.Type = &typ
	}

	if qp.StartCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.StartCreatedDate)
		if err != nil {
			return homebus.QueryFilter{}, errs.NewFieldsError("start_created_date", err)
		}
		filter.StartCreatedDate = &t
	}

	if qp.EndCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.EndCreatedDate)
		if err != nil {
			return homebus.QueryFilter{}, errs.NewFieldsError("end_created_date", err)
		}
		filter.EndCreatedDate = &t
	}

	return filter, nil
}
