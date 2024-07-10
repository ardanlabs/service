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

// Respond sends a response to the client.
func Respond(ctx context.Context, w http.ResponseWriter, dataModel Encoder) error {

	// If the context has been canceled, it means the client is no longer
	// waiting for a response.
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			return errors.New("client disconnected, do not send response")
		}
	}

	var statusCode = http.StatusOK

	switch v := dataModel.(type) {
	case httpStatus:
		statusCode = v.HTTPStatus()

	case error:
		statusCode = http.StatusInternalServerError

	default:
		if dataModel == nil {
			statusCode = http.StatusNoContent
		}
	}

	_, span := tracer.AddSpan(ctx, "foundation.response", attribute.Int("status", statusCode))
	defer span.End()

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	data, contentType, err := dataModel.Encode()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("respond: encode: %w", err)
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("respond: write: %w", err)
	}

	return nil
}
