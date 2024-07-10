package web

import (
	"context"
	"net/http"
)

type ctxKey int

const (
	traceKey ctxKey = iota + 1
	writer
	logger
)

func setTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceKey, traceID)
}

// GetTraceID returns the trace id from the context.
func GetTraceID(ctx context.Context) string {
	v, ok := ctx.Value(traceKey).(string)
	if !ok {
		return "00000000-0000-0000-0000-000000000000"
	}

	return v
}

func setWriter(ctx context.Context, w http.ResponseWriter) context.Context {
	return context.WithValue(ctx, writer, w)
}

func getWriter(ctx context.Context) http.ResponseWriter {
	v, ok := ctx.Value(writer).(http.ResponseWriter)
	if !ok {
		return nil
	}

	return v
}

func setLogger(ctx context.Context, log Logger) context.Context {
	return context.WithValue(ctx, writer, log)
}

func getLogger(ctx context.Context) Logger {
	v, ok := ctx.Value(logger).(Logger)
	if !ok {
		return func(ctx context.Context, msg string, args ...any) {}
	}

	return v
}
