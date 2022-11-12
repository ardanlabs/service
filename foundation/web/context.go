package web

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ctxKey int

const key ctxKey = 1

// Values represent state for each request.
type Values struct {
	TraceID    string
	Tracer     trace.Tracer
	Now        time.Time
	StatusCode int
}

// GetValues returns the values from the context.
func GetValues(ctx context.Context) *Values {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return &Values{
			TraceID: "00000000-0000-0000-0000-000000000000",
			Tracer:  trace.NewNoopTracerProvider().Tracer(""),
			Now:     time.Now(),
		}
	}

	return v
}

// GetTraceID returns the trace id from the context.
func GetTraceID(ctx context.Context) string {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return "00000000-0000-0000-0000-000000000000"
	}
	return v.TraceID
}

// GetTime returns the time from the context.
func GetTime(ctx context.Context) time.Time {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return time.Now()
	}
	return v.Now
}

// AddSpan adds a OpenTelemetry span to the trace and context.
func AddSpan(ctx context.Context, spanName string, keyValues ...attribute.KeyValue) (context.Context, trace.Span) {
	v, ok := ctx.Value(key).(*Values)
	if !ok || v.Tracer == nil {
		return ctx, trace.SpanFromContext(ctx)
	}

	ctx, span := v.Tracer.Start(ctx, spanName)
	for _, kv := range keyValues {
		span.SetAttributes(kv)
	}

	return ctx, span
}

// SetStatusCode sets the status code back into the context.
func SetStatusCode(ctx context.Context, statusCode int) {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return
	}

	v.StatusCode = statusCode
}
