package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/foundation/tracer"
	"github.com/go-json-experiment/json"
	"go.opentelemetry.io/otel/attribute"
)

type Response struct {
	Err        error
	Data       any
	StatusCode int
}

func EmptyResponse() Response {
	return Response{}
}

func Respond(data any, statusCode int) Response {
	return Response{
		Data:       data,
		StatusCode: statusCode,
	}
}

// Respond constructs an error reponse value.
func RespondError(err error, statusCode int) Response {
	return Response{
		Err:        err,
		StatusCode: statusCode,
	}
}

func (r Response) send(ctx context.Context, w http.ResponseWriter) error {
	_, span := tracer.AddSpan(ctx, "foundation.web.response", attribute.Int("status", r.StatusCode))
	defer span.End()

	if r.StatusCode == http.StatusNoContent {
		w.WriteHeader(r.StatusCode)
		return nil
	}

	var data any

	switch {
	case r.Err != nil:
		data = r.Err

	default:
		data = r.Data
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("web.respond: marshal: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.StatusCode)

	if _, err := w.Write(jsonData); err != nil {
		return fmt.Errorf("web.respond: write: %w", err)
	}

	return nil
}
