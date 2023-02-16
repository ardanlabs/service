package usergrp

import (
	"net/http"
	"net/mail"

	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/sys/validate"
	"github.com/google/uuid"
)

func parseFilter(r *http.Request) (user.QueryFilter, error) {
	values := r.URL.Query()

	var filter user.QueryFilter

	if id, err := uuid.Parse(values.Get("id")); err == nil {
		filter.ByID(id)
	}

	if err := filter.ByName(values.Get("name")); err != nil {
		return user.QueryFilter{}, validate.NewFieldsError("name", err)
	}

	if email, err := mail.ParseAddress(values.Get("email")); err == nil {
		filter.ByEmail(*email)
	}

	if err := filter.Validate(); err != nil {
		return user.QueryFilter{}, err
	}

	return filter, nil
}
