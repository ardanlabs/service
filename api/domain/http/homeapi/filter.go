package homeapi

import (
	"net/http"

	"github.com/ardanlabs/service/app/domain/homeapp"
)

func parseQueryParams(r *http.Request) (homeapp.QueryParams, error) {
	values := r.URL.Query()

	filter := homeapp.QueryParams{
		Page:             values.Get("page"),
		Rows:             values.Get("row"),
		OrderBy:          values.Get("orderBy"),
		ID:               values.Get("home_id"),
		UserID:           values.Get("user_id"),
		Type:             values.Get("type"),
		StartCreatedDate: values.Get("start_created_date"),
		EndCreatedDate:   values.Get("end_created_date"),
	}

	return filter, nil
}
