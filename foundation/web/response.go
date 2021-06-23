package web

import (
	"context"
	"encoding/json"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// Respond converts a Go value to JSON and sends it to the client.
func Respond(ctx context.Context, w http.ResponseWriter, data interface{}, statusCode int) error {
	ctx, span := otel.GetTracerProvider().Tracer("").Start(ctx, "foundation.web.respond")
	span.SetAttributes(attribute.Int("statusCode", statusCode))
	defer span.End()

	// Set the status code for the request logger middleware.
	// If the context is missing this value, don't set it and
	// make sure a response is provided.
	if v, ok := ctx.Value(KeyValues).(*Values); ok {
		v.StatusCode = statusCode
	}

	// If there is nothing to marshal then set status code and return.
	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	// Convert the response value to JSON.
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Set the content type and headers once we know marshaling has succeeded.
	w.Header().Set("Content-Type", "application/json")

	// Write the status code to the response.
	w.WriteHeader(statusCode)

	// Send the result back to the client.
	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}
