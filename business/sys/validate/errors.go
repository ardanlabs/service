package validate

import (
	"encoding/json"
	"errors"
)

// FieldError is used to indicate an error with a specific request field.
type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

// FieldErrors represents a collection of field errors.
type FieldErrors []FieldError

// Error implments the error interface.
func (fe FieldErrors) Error() string {
	d, err := json.Marshal(fe)
	if err != nil {
		return err.Error()
	}
	return string(d)
}

// AsFieldErrors checks if an error of type FieldErrors exists.
func AsFieldErrors(err error) bool {
	var fieldErrors FieldErrors
	return errors.As(err, &fieldErrors)
}

// GetFieldErrors returns a copy of the FieldErrors pointer.
func GetFieldErrors(err error) FieldErrors {
	var fieldErrors FieldErrors
	if !errors.As(err, &fieldErrors) {
		return nil
	}
	return fieldErrors
}
