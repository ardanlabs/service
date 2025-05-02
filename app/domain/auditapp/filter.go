package auditapp

import (
	"net/http"
	"time"

	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/business/domain/auditbus"
	"github.com/ardanlabs/service/business/types/domain"
	"github.com/ardanlabs/service/business/types/name"
	"github.com/google/uuid"
)

type queryParams struct {
	Page      string
	Rows      string
	OrderBy   string
	ObjID     string
	ObjDomain string
	ObjName   string
	ActorID   string
	Action    string
	Since     string
	Until     string
}

func parseQueryParams(r *http.Request) (queryParams, error) {
	values := r.URL.Query()

	filter := queryParams{
		Page:      values.Get("page"),
		Rows:      values.Get("rows"),
		OrderBy:   values.Get("orderBy"),
		ObjID:     values.Get("obj_id"),
		ObjDomain: values.Get("obj_domain"),
		ObjName:   values.Get("obj_name"),
		ActorID:   values.Get("actor_id"),
		Action:    values.Get("action"),
		Since:     values.Get("since"),
		Until:     values.Get("until"),
	}

	return filter, nil
}

func parseFilter(qp queryParams) (auditbus.QueryFilter, error) {
	var fieldErrors errs.FieldErrors
	var filter auditbus.QueryFilter

	if qp.ObjID != "" {
		id, err := uuid.Parse(qp.ObjID)
		switch err {
		case nil:
			filter.ObjID = &id
		default:
			fieldErrors.Add("obj_id", err)
		}
	}

	if qp.ObjDomain != "" {
		domain, err := domain.Parse(qp.ObjDomain)
		switch err {
		case nil:
			filter.ObjDomain = &domain
		default:
			fieldErrors.Add("obj_domain", err)
		}
	}

	if qp.ObjName != "" {
		name, err := name.Parse(qp.ObjName)
		switch err {
		case nil:
			filter.ObjName = &name
		default:
			fieldErrors.Add("obj_name", err)
		}
	}

	if qp.ActorID != "" {
		id, err := uuid.Parse(qp.ActorID)
		switch err {
		case nil:
			filter.ActorID = &id
		default:
			fieldErrors.Add("actor_id", err)
		}
	}

	if qp.Action != "" {
		filter.Action = &qp.Action
	}

	if qp.Since != "" {
		t, err := time.Parse(time.RFC3339, qp.Since)
		switch err {
		case nil:
			filter.Since = &t
		default:
			fieldErrors.Add("since", err)
		}
	}

	if qp.Until != "" {
		t, err := time.Parse(time.RFC3339, qp.Until)
		switch err {
		case nil:
			filter.Until = &t
		default:
			fieldErrors.Add("until", err)
		}
	}

	if fieldErrors != nil {
		return auditbus.QueryFilter{}, fieldErrors.ToError()
	}

	return filter, nil
}
