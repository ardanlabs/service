package userapp

import (
	"net/mail"
	"time"

	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/foundation/validate"
	"github.com/google/uuid"
)

func parseFilter(qp QueryParams) (user.QueryFilter, error) {
	var filter user.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return user.QueryFilter{}, validate.NewFieldsError("user_id", err)
		}
		filter.WithUserID(id)
	}

	if qp.Name != "" {
		filter.WithName(qp.Name)
	}

	if qp.Email != "" {
		addr, err := mail.ParseAddress(qp.Email)
		if err != nil {
			return user.QueryFilter{}, validate.NewFieldsError("email", err)
		}
		filter.WithEmail(*addr)
	}

	if qp.StartCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.StartCreatedDate)
		if err != nil {
			return user.QueryFilter{}, validate.NewFieldsError("start_created_date", err)
		}
		filter.WithStartDateCreated(t)
	}

	if qp.EndCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.EndCreatedDate)
		if err != nil {
			return user.QueryFilter{}, validate.NewFieldsError("end_created_date", err)
		}
		filter.WithEndCreatedDate(t)
	}

	return filter, nil
}
