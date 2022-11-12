package auth

import (
	"errors"
	"fmt"
)

// AuthError is used to pass an error during the request through the
// application with auth specific context.
type AuthError struct {
	msg string
}

// NewAuthError wraps a provided error with an HTTP status code of
// StatusForbidden.
func NewAuthError(err error) error {
	return &AuthError{
		msg: err.Error(),
	}
}

// NewAuthErrorf wraps a provided error with an HTTP status code of
// StatusForbidden.
func NewAuthErrorf(format string, args ...any) error {
	return &AuthError{
		msg: fmt.Sprintf(format, args...),
	}
}

// Error implements the error interface. It uses the default message of the
// wrapped error. This is what will be shown in the services' logs.
func (ae *AuthError) Error() string {
	return ae.msg
}

// IsAuthError checks if an error of type AuthError exists.
func IsAuthError(err error) bool {
	var ae *AuthError
	return errors.As(err, &ae)
}
