package homegrp

import (
	"net/http"
	"time"

	"github.com/ardanlabs/service/business/core/crud/home"
	"github.com/ardanlabs/service/foundation/validate"
	"github.com/google/uuid"
)

func parseFilter(r *http.Request) (home.QueryFilter, error) {
	const (
		filterByHomeID           = "home_id"
		filterByUserID           = "user_id"
		filterByType             = "type"
		filterByStartCreatedDate = "start_date_created"
		filterByEndCreatedDate   = "end_date_created"
	)

	values := r.URL.Query()

	var filter home.QueryFilter

	if homeID := values.Get(filterByHomeID); homeID != "" {
		id, err := uuid.Parse(homeID)
		if err != nil {
			return home.QueryFilter{}, validate.NewFieldsError(filterByHomeID, err)
		}
		filter.WithHomeID(id)
	}

	if userID := values.Get(filterByUserID); userID != "" {
		id, err := uuid.Parse(userID)
		if err != nil {
			return home.QueryFilter{}, validate.NewFieldsError(filterByUserID, err)
		}
		filter.WithUserID(id)
	}

	if homeType := values.Get(filterByType); homeType != "" {
		typ, err := home.ParseType(homeType)
		if err != nil {
			return home.QueryFilter{}, validate.NewFieldsError(filterByType, err)
		}
		filter.WithHomeType(typ)
	}

	if createdDate := values.Get(filterByStartCreatedDate); createdDate != "" {
		t, err := time.Parse(time.RFC3339, createdDate)
		if err != nil {
			return home.QueryFilter{}, validate.NewFieldsError(filterByStartCreatedDate, err)
		}
		filter.WithStartDateCreated(t)
	}

	if createdDate := values.Get(filterByEndCreatedDate); createdDate != "" {
		t, err := time.Parse(time.RFC3339, createdDate)
		if err != nil {
			return home.QueryFilter{}, validate.NewFieldsError(filterByEndCreatedDate, err)
		}
		filter.WithEndCreatedDate(t)
	}

	return filter, nil
}
