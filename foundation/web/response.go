package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/foundation/tracer"
	"go.opentelemetry.io/otel/attribute"
)

type httpStatus interface {
	HTTPStatus() int
}

func respondError(ctx context.Context, w http.ResponseWriter, err error) error {
	data, ok := err.(Encoder)
	if !ok {
		return errors.New("error value does not implement the encoder interface")
	}

	return respond(ctx, w, data)
}

func respond(ctx context.Context, w http.ResponseWriter, data Encoder) error {
	var statusCode = http.StatusOK
	switch v := data.(type) {
	case httpStatus:
		statusCode = v.HTTPStatus()
	case error:
		statusCode = http.StatusInternalServerError
	}

	_, span := tracer.AddSpan(ctx, "foundation.response", attribute.Int("status", statusCode))
	defer span.End()

	if data == nil {
		statusCode = http.StatusNoContent
	}

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	b, err := data.Encode()
	if err != nil {
		return fmt.Errorf("respond: encode: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if _, err := w.Write(b); err != nil {
		return fmt.Errorf("respond: write: %w", err)
	}

	return nil
}
