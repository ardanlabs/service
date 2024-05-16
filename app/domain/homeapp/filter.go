package homeapp

import (
	"time"

	"github.com/ardanlabs/service/business/domain/homebus"
	"github.com/ardanlabs/service/foundation/validate"
	"github.com/google/uuid"
)

func parseFilter(qp QueryParams) (homebus.QueryFilter, error) {
	var filter homebus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return homebus.QueryFilter{}, validate.NewFieldsError("home_id", err)
		}
		filter.ID = &id
	}

	if qp.UserID != "" {
		id, err := uuid.Parse(qp.UserID)
		if err != nil {
			return homebus.QueryFilter{}, validate.NewFieldsError("user_id", err)
		}
		filter.UserID = &id
	}

	if qp.Type != "" {
		typ, err := homebus.Types.Parse(qp.Type)
		if err != nil {
			return homebus.QueryFilter{}, validate.NewFieldsError("type", err)
		}
		filter.Type = &typ
	}

	if qp.StartCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.StartCreatedDate)
		if err != nil {
			return homebus.QueryFilter{}, validate.NewFieldsError("start_created_date", err)
		}
		filter.StartCreatedDate = &t
	}

	if qp.EndCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.EndCreatedDate)
		if err != nil {
			return homebus.QueryFilter{}, validate.NewFieldsError("end_created_date", err)
		}
		filter.EndCreatedDate = &t
	}

	return filter, nil
}
