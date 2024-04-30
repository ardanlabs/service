package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/foundation/tracer"
	"github.com/go-json-experiment/json"
	"go.opentelemetry.io/otel/attribute"
)

// Respond converts a Go value to JSON and sends it to the client.
func Respond(ctx context.Context, w http.ResponseWriter, data any, statusCode int) error {
	ctx, span := tracer.AddSpan(ctx, "foundation.web.response", attribute.Int("status", statusCode))
	defer span.End()

	setStatusCode(ctx, statusCode)

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
