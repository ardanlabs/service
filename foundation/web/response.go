package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/foundation/tracer"
	"github.com/go-json-experiment/json"
	"go.opentelemetry.io/otel/attribute"
)

type httpStatus interface {
	HTTPStatus() int
}

func respond(ctx context.Context, w http.ResponseWriter, data any) error {
	var statusCode = http.StatusOK
	if _, ok := data.(error); ok {
		if v, ok := data.(httpStatus); ok {
			statusCode = v.HTTPStatus()
		} else {
			statusCode = http.StatusInternalServerError
		}
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
