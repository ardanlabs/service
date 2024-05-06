package web

import (
	"fmt"
	"io"
	"net/http"
)

// Param returns the web call parameters from the request.
func Param(r *http.Request, key string) string {
	return r.PathValue(key)
}

type decoder interface {
	Decode(data []byte) error
}

type validator interface {
	Validate() error
}

// Decode reads the body of an HTTP request. If the data model provided
// implements the decoder interface, that implemementation is used to decode
// the body. If the interface is not implemented, an error is returned.
func Decode(r *http.Request, val any) error {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("web.request: unable to read payload: %w", err)
	}

	dec, ok := val.(decoder)
	if !ok {
		return fmt.Errorf("web.request: encoder not implemented")
	}

	if err := dec.Decode(data); err != nil {
		return fmt.Errorf("web.request: encode: %w", err)
	}

	if v, ok := val.(validator); ok {
		if err := v.Validate(); err != nil {
			return err
		}
	}

	return nil
}
