package web

import (
	"fmt"
	"net/http"

	"github.com/go-json-experiment/json"
)

// Param returns the web call parameters from the request.
func Param(r *http.Request, key string) string {
	return r.PathValue(key)
}

type validator interface {
	Validate() error
}

// Decode reads the body of an HTTP request looking for a JSON document. The
// body is decoded into the provided value.
// If the provided value is a struct then it is checked for validation tags.
// If the value implements a validate function, it is executed.
func Decode(r *http.Request, val any) error {
	if err := json.UnmarshalRead(r.Body, val, json.RejectUnknownMembers(false)); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	if v, ok := val.(validator); ok {
		if err := v.Validate(); err != nil {
			return err
		}
	}

	return nil
}
