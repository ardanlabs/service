// Package errs provides types and support related to web error functionality.
package errs

import "errors"

// Response is the form used for API responses from failures in the API.
type Response struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

// Trusted is used to pass an error during the request through the
// application with web specific context.
type Trusted struct {
	Err    error
	Status int
}

// NewTrusted wraps a provided error with an HTTP status code. This
// function should be used when handlers encounter expected errors.
func NewTrusted(err error, status int) error {
	return &Trusted{err, status}
}

// Error implements the error interface. It uses the default message of the
// wrapped error. This is what will be shown in the services' logs.
func (te *Trusted) Error() string {
	return te.Err.Error()
}

// IsTrusted checks if an error of type TrustedError exists.
func IsTrusted(err error) bool {
	var te *Trusted
	return errors.As(err, &te)
}

// GetTrusted returns a copy of the TrustedError pointer.
func GetTrusted(err error) *Trusted {
	var te *Trusted
	if !errors.As(err, &te) {
		return nil
	}
	return te
}
