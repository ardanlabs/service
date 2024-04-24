// Package errs provides types and support related to web error functionality.
package errs

import (
	"errors"
	"fmt"
)

// Error represents an error in the system.
type Error struct {
	Code    ErrCode `json:"code"`
	Message string  `json:"message"`
}

// New constructs an error based on an app error.
func New(code ErrCode, err error) Error {
	return Error{
		Code:    code,
		Message: err.Error(),
	}
}

// Newf constructs an error based on a error message.
func Newf(code ErrCode, format string, v ...any) Error {
	return Error{
		Code:    code,
		Message: fmt.Sprintf(format, v...),
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
