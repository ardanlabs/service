package userapi

import (
	"net/http"

	"github.com/ardanlabs/service/app/domain/userapp"
)

func parseQueryParams(r *http.Request) (userapp.QueryParams, error) {
	values := r.URL.Query()

	filter := userapp.QueryParams{
		Page:             values.Get("page"),
		Rows:             values.Get("row"),
		OrderBy:          values.Get("orderBy"),
		ID:               values.Get("user_id"),
		Name:             values.Get("name"),
		Email:            values.Get("email"),
		StartCreatedDate: values.Get("start_created_date"),
		EndCreatedDate:   values.Get("end_created_date"),
	}

	return filter, nil
}
