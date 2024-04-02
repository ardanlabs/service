// Package errs provides types and support related to web error functionality.
package errs

import (
	"errors"
	"fmt"
	"net/http"
)

var errMap = map[int]ErrCode{
	http.StatusOK:                  OK,
	http.StatusInternalServerError: Internal,
	http.StatusBadRequest:          FailedPrecondition,
	http.StatusGatewayTimeout:      DeadlineExceeded,
	http.StatusNotFound:            NotFound,
	http.StatusConflict:            Aborted,
	http.StatusForbidden:           PermissionDenied,
	http.StatusTooManyRequests:     ResourceExhausted,
	http.StatusNotImplemented:      Unimplemented,
	http.StatusServiceUnavailable:  Unavailable,
	http.StatusUnauthorized:        Unauthenticated,
}

// Details provides the caller with more error context.
type Details struct {
	HTTPStatusCode int    `json:"httpStatusCode"`
	HTTPStatus     string `json:"httpStatus"`
}

// Error represents an error in the system.
type Error struct {
	Code    ErrCode `json:"code"`
	Message string  `json:"message"`
	Details Details `json:"details"`
}

// New constructs an error based on an app error.
func New(httpStatus int, err error) Error {
	return Error{
		Code:    errMap[httpStatus],
		Message: err.Error(),
		Details: Details{
			HTTPStatusCode: httpStatus,
			HTTPStatus:     http.StatusText(httpStatus),
		},
	}
}

// Newf constructs an error based on a error message.
func Newf(httpStatus int, format string, v ...any) Error {
	return Error{
		Code:    errMap[httpStatus],
		Message: fmt.Sprintf(format, v...),
		Details: Details{
			HTTPStatusCode: httpStatus,
			HTTPStatus:     http.StatusText(httpStatus),
		},
	}
}

// Error implements the error interface.
func (err Error) Error() string {
	return err.Message
}

// IsError tests the concrete error is of the Error type.
func IsError(err error) bool {
	var er Error
	return errors.As(err, &er)
}

// GetError returns a copy of the Error.
func GetError(err error) Error {
	var er Error
	if !errors.As(err, &er) {
		return Error{}
	}
	return er
}
