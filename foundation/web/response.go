package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/foundation/tracer"
	"go.opentelemetry.io/otel/attribute"
)

type httpStatus interface {
	HTTPStatus() int
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

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("web.respond: marshal: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if _, err := w.Write(jsonData); err != nil {
		return fmt.Errorf("web.respond: write: %w", err)
	}

	return nil
}
