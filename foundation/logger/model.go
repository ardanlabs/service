package logger

import (
	"context"
	"time"

	"log/slog"
)

// Level represents different logging levels.
type Level slog.Level

// A set of possible logging levels.
const (
	LevelDebug = Level(slog.LevelDebug)
	LevelInfo  = Level(slog.LevelInfo)
	LevelWarn  = Level(slog.LevelWarn)
	LevelError = Level(slog.LevelError)
)

// =============================================================================

// Record represents the data that is being logged.
type Record struct {
	Time       time.Time
	Message    string
	Level      Level
	Attributes map[string]any
}

func toRecord(r slog.Record) Record {
	atts := make(map[string]any, r.NumAttrs())

	f := func(attr slog.Attr) bool {
		atts[attr.Key] = attr.Value.Any()
		return true
	}
	r.Attrs(f)

	return Record{
		Time:       r.Time,
		Message:    r.Message,
		Level:      Level(r.Level),
		Attributes: atts,
	}
}

// EventFunc is a function to be executed when configured against a log level.
type EventFunc func(ctx context.Context, r Record)

// Events contains an assignment of an event function to a log level.
type Events struct {
	Debug EventFunc
	Info  EventFunc
	Warn  EventFunc
	Error EventFunc
}
