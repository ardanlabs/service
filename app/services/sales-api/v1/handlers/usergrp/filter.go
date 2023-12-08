package usergrp

import (
	"net/http"
	"net/mail"
	"time"

	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/foundation/validate"
	"github.com/google/uuid"
)

func parseFilter(r *http.Request) (user.QueryFilter, error) {
	const (
		filterByUserID           = "user_id"
		filterByEmail            = "email"
		filterByStartCreatedDate = "start_created_date"
		filterByEndCreatedDate   = "end_created_date"
		filterByName             = "name"
	)

	values := r.URL.Query()

	var filter user.QueryFilter

	if userID := values.Get(filterByUserID); userID != "" {
		id, err := uuid.Parse(userID)
		if err != nil {
			return user.QueryFilter{}, validate.NewFieldsError(filterByUserID, err)
		}
		filter.WithUserID(id)
	}

	if email := values.Get(filterByEmail); email != "" {
		addr, err := mail.ParseAddress(email)
		if err != nil {
			return user.QueryFilter{}, validate.NewFieldsError(filterByEmail, err)
		}
		filter.WithEmail(*addr)
	}

	if createdDate := values.Get(filterByStartCreatedDate); createdDate != "" {
		t, err := time.Parse(time.RFC3339, createdDate)
		if err != nil {
			return user.QueryFilter{}, validate.NewFieldsError(filterByStartCreatedDate, err)
		}
		filter.WithStartDateCreated(t)
	}

	if createdDate := values.Get(filterByEndCreatedDate); createdDate != "" {
		t, err := time.Parse(time.RFC3339, createdDate)
		if err != nil {
			return user.QueryFilter{}, validate.NewFieldsError(filterByEndCreatedDate, err)
		}
		filter.WithEndCreatedDate(t)
	}

	if name := values.Get(filterByName); name != "" {
		filter.WithName(name)
	}

	return filter, nil
}
