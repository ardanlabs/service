package otel

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type ctxKey int

const (
	tracerKey ctxKey = iota + 1
	traceIDKey
)

func setTracer(ctx context.Context, tracer trace.Tracer) context.Context {
	return context.WithValue(ctx, tracerKey, tracer)
}

func setTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID returns the trace id from the context.
func GetTraceID(ctx context.Context) string {
	v, ok := ctx.Value(traceIDKey).(string)
	if !ok {
		return defaultTraceID
	}

	return v
}
