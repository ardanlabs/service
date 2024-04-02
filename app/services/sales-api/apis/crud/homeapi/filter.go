package homeapi

import (
	"time"

	"github.com/ardanlabs/service/business/core/crud/home"
	"github.com/ardanlabs/service/foundation/validate"
	"github.com/google/uuid"
)

func parseFilter(qp QueryParams) (home.QueryFilter, error) {
	var filter home.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return home.QueryFilter{}, validate.NewFieldsError("home_id", err)
		}
		filter.WithHomeID(id)
	}

	if qp.UserID != "" {
		id, err := uuid.Parse(qp.UserID)
		if err != nil {
			return home.QueryFilter{}, validate.NewFieldsError("user_id", err)
		}
		filter.WithUserID(id)
	}

	if qp.Type != "" {
		typ, err := home.ParseType(qp.Type)
		if err != nil {
			return home.QueryFilter{}, validate.NewFieldsError("type", err)
		}
		filter.WithHomeType(typ)
	}

	if qp.StartCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.StartCreatedDate)
		if err != nil {
			return home.QueryFilter{}, validate.NewFieldsError("start_created_date", err)
		}
		filter.WithStartDateCreated(t)
	}

	if qp.EndCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, qp.EndCreatedDate)
		if err != nil {
			return home.QueryFilter{}, validate.NewFieldsError("end_created_date", err)
		}
		filter.WithEndCreatedDate(t)
	}

	return filter, nil
}
