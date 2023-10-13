// Package response provides the response and error handling types support.
package response

import (
	"errors"
)

// PageDocument is the form used for API responses from query API calls.
type PageDocument[T any] struct {
	Items       []T `json:"items"`
	Total       int `json:"total"`
	Page        int `json:"page"`
	RowsPerPage int `json:"rowsPerPage"`
}

// NewPageDocument constructs a response value for a web paging response.
func NewPageDocument[T any](items []T, total int, page int, rowsPrePage int) PageDocument[T] {
	return PageDocument[T]{
		Items:       items,
		Total:       total,
		Page:        page,
		RowsPerPage: rowsPrePage,
	}
}

// =============================================================================

// ErrorDocument is the form used for API responses from failures in the API.
type ErrorDocument struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

// Error is used to pass an error during the request through the
// application with web specific context.
type Error struct {
	Err    error
	Status int
}

// NewError wraps a provided error with an HTTP status code. This
// function should be used when handlers encounter expected errors.
func NewError(err error, status int) error {
	return &Error{err, status}
}

// Error implements the error interface. It uses the default message of the
// wrapped error. This is what will be shown in the services' logs.
func (re *Error) Error() string {
	return re.Err.Error()
}

// IsError checks if an error of type Error exists.
func IsError(err error) bool {
	var re *Error
	return errors.As(err, &re)
}

// GetError returns a copy of the Error pointer.
func GetError(err error) *Error {
	var re *Error
	if !errors.As(err, &re) {
		return nil
	}
	return re
}
