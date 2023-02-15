// Package v1 represents types used by the web application for v1.
package v1

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ardanlabs/service/business/data/order"
	"github.com/ardanlabs/service/business/sys/validate"
)

// ErrorResponse is the form used for API responses from failures in the API.
type ErrorResponse struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

// RequestError is used to pass an error during the request through the
// application with web specific context.
type RequestError struct {
	Err    error
	Status int
}

// NewRequestError wraps a provided error with an HTTP status code. This
// function should be used when handlers encounter expected errors.
func NewRequestError(err error, status int) error {
	return &RequestError{err, status}
}

// Error implements the error interface. It uses the default message of the
// wrapped error. This is what will be shown in the services' logs.
func (re *RequestError) Error() string {
	return re.Err.Error()
}

// IsRequestError checks if an error of type RequestError exists.
func IsRequestError(err error) bool {
	var re *RequestError
	return errors.As(err, &re)
}

// GetRequestError returns a copy of the RequestError pointer.
func GetRequestError(err error) *RequestError {
	var re *RequestError
	if !errors.As(err, &re) {
		return nil
	}
	return re
}

// =============================================================================

// ParseOrderBy constructs an order.By value by parsing a string in the form
// of "field,direction" from the request.
func ParseOrderBy(r *http.Request, orderingFields order.FieldSet, defaultOrder order.By) (order.By, error) {
	v := r.URL.Query().Get("orderBy")

	if v == "" {
		return defaultOrder, nil
	}

	orderParts := strings.Split(v, ",")

	var by order.By
	switch len(orderParts) {
	case 1:
		field, err := orderingFields.ParseField(strings.Trim(orderParts[0], " "))
		if err != nil {
			return order.By{}, validate.NewFieldsError(v, errors.New("parsing fields"))
		}

		by = order.NewBy(field, order.ASC)

	case 2:
		field, err := orderingFields.ParseField(strings.Trim(orderParts[0], " "))
		if err != nil {
			return order.By{}, validate.NewFieldsError(v, errors.New("parsing fields"))
		}

		dir, err := order.ParseDirection(strings.Trim(orderParts[1], " "))
		if err != nil {
			return order.By{}, validate.NewFieldsError(v, errors.New("parsing direction"))
		}

		by = order.NewBy(field, dir)

	default:
		return order.By{}, validate.NewFieldsError(v, errors.New("unknown order field"))
	}

	return by, nil
}
