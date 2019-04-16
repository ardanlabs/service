package web

import (
	"net/http"

	"github.com/pkg/errors"
)

// FieldError is used to indicate an error with a specific request field.
type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

// ErrorResponse is the form used for API responses from failures in the API.
type ErrorResponse struct {
	Error  string       `json:"error"`
	Fields []FieldError `json:"fields,omitempty"`
}

// StatusError is used to pass errors through the application with web specific
// context.
type StatusError struct {
	err    error
	Status int
	Fields []FieldError
}

// WrapErrorWithStatus wraps a provided error with an HTTP status code. This
// function should be used when handlers encounter expected errors.
func WrapErrorWithStatus(err error, status int) error {
	return &StatusError{err, status, nil}
}

// Error implements the error interface. It uses the default message of the
// wrapped error. This is what will be shown in the services' logs.
func (se *StatusError) Error() string {
	return se.err.Error()
}

// String provides "human readable" error messages that are intended for
// service users to see. If the status code is 500 or higher (the default) then
// a generic error message is returned.
//
// The idea is that a developer who creates an error like this intends to let
// the API consumer know the product was not found:
//	WrapErrorWithStatus(errors.New("product not found"), 404)
//
// However a more serious error like a database failure might include
// information that is not safe to show to API consumers.
func (se *StatusError) String() string {
	if se.Status < http.StatusInternalServerError {
		return se.err.Error()
	}
	return http.StatusText(se.Status)
}

// NewStatusError takes a regular error and converts it to a StatusError. If the
// original error is already a *StatusError it is returned directly. If not
// then it is defaulted to an error with a 500 status.
func NewStatusError(err error) *StatusError {

	// errors.Cause() returns the original error that may have been wrapped by
	// the errors package. The return value from that function is the error
	// interface type so this code runs a type assertion to see if that original
	// error was of the type *StatusError. If it was then it is returned.
	if se, ok := errors.Cause(err).(*StatusError); ok {
		return se
	}

	// If the original error was NOT a *StatusError then an error with a default
	// 500 status code is returned.
	return &StatusError{err, http.StatusInternalServerError, nil}
}

// shutdown is a type used to help with the graceful termination of the service.
type shutdown struct {
	Message string
}

// Error is the implementation of the error interface.
func (s *shutdown) Error() string {
	return s.Message
}

// Shutdown returns an error that causes the framework to signal
// a graceful shutdown.
func Shutdown(message string) error {
	return &shutdown{message}
}

// IsShutdown checks to see if the shutdown error is contained
// in the specified error value.
func IsShutdown(err error) bool {
	if _, ok := errors.Cause(err).(*shutdown); ok {
		return true
	}
	return false
}
