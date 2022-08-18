package web

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// ctxKey represents the type of value for the context key.
type ctxKey int

// key is how request values are stored/retrieved.
const key ctxKey = 1

// Values represent state for each request.
type Values struct {
	TraceID    string
	Tracer     trace.Tracer
	Now        time.Time
	StatusCode int
}

// GetValues returns the values from the context.
func GetValues(ctx context.Context) (*Values, error) {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return nil, errors.New("web value missing from context")
	}
	return v, nil
}

// GetTraceID returns the trace id from the context.
func GetTraceID(ctx context.Context) string {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return "00000000-0000-0000-0000-000000000000"
	}
	return v.TraceID
}

// GetTracer returns the tracer from the context.
func GetTracer(ctx context.Context) (trace.Tracer, error) {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return nil, errors.New("tracer does not exist")
	}
	return v.Tracer, nil
}

// SetStatusCode sets the status code back into the context.
func SetStatusCode(ctx context.Context, statusCode int) error {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return errors.New("web value missing from context")
	}
	v.StatusCode = statusCode
	return nil
}
