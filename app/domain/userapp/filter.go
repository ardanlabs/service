package userapp

import (
	"net/mail"
	"time"

	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/foundation/validate"
	"github.com/google/uuid"
)

func parseFilter(qp QueryParams) (userbus.QueryFilter, error) {
	var filter userbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return userbus.QueryFilter{}, validate.NewFieldsError("user_id", err)
		}
		filter.ID = &id
	}

	if qp.Name != "" {
		name, err := userbus.Names.Parse(qp.Name)
		if err != nil {
			return userbus.QueryFilter{}, validate.NewFieldsError("name", err)
		}
		filter.Name = &name
	}

	if qp.Email != "" {
		addr, err := mail.ParseAddress(qp.Email)
		if err != nil {
			return userbus.QueryFilter{}, validate.NewFieldsError("email", err)
		}
		filter.Email = addr
	}

	if qp.StartCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.StartCreatedDate)
		if err != nil {
			return userbus.QueryFilter{}, validate.NewFieldsError("start_created_date", err)
		}
		filter.StartCreatedDate = &t
	}

	if qp.EndCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.EndCreatedDate)
		if err != nil {
			return userbus.QueryFilter{}, validate.NewFieldsError("end_created_date", err)
		}
		filter.EndCreatedDate = &t
	}

	return filter, nil
}
