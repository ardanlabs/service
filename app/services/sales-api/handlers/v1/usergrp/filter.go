package usergrp

import (
	"net/http"

	"github.com/ardanlabs/service/business/core/user"
)

func getFilter(r *http.Request) (user.QueryFilter, error) {
	values := r.URL.Query()

	var filter user.QueryFilter
	filter.ByID(values.Get("id"))
	filter.ByName(values.Get("name"))
	filter.ByEmail(values.Get("email"))

	return filter, nil
}
