package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/foundation/tracer"
	"go.opentelemetry.io/otel/attribute"
)

type httpStatus interface {
	HTTPStatus() int
}

type encoder interface {
	Encode() ([]byte, error)
}

func respond(ctx context.Context, w http.ResponseWriter, data any) error {
	var statusCode = http.StatusOK
	switch v := data.(type) {
	case httpStatus:
		statusCode = v.HTTPStatus()
	case error:
		statusCode = http.StatusInternalServerError
	}

	_, span := tracer.AddSpan(ctx, "foundation.web.response", attribute.Int("status", statusCode))
	defer span.End()

	if data == nil {
		statusCode = http.StatusNoContent
	}

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	enc, ok := data.(encoder)
	if !ok {
		return fmt.Errorf("web.respond: encoder not implemented")
	}

	b, err := enc.Encode()
	if err != nil {
		return fmt.Errorf("web.respond: encode: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if _, err := w.Write(b); err != nil {
		return fmt.Errorf("web.respond: write: %w", err)
	}

	return nil
}
