package usergrp

import (
	"net/http"
	"net/mail"

	"github.com/ardanlabs/service/business/core/user"
	"github.com/google/uuid"
)

func getFilter(r *http.Request) (user.QueryFilter, error) {
	values := r.URL.Query()

	var filter user.QueryFilter
	if id, err := uuid.Parse(values.Get("id")); err == nil {
		filter.ByID(id)
	}

	filter.ByName(values.Get("name"))

	if email, err := mail.ParseAddress(values.Get("email")); err == nil {
		filter.ByEmail(*email)
	}

	return filter, nil
}
