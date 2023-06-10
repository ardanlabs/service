// Package logger provides support for initializing the log system.
package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/exp/slog"
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
	Time    time.Time
	Message string
	Level   Level
}

func toRecord(r slog.Record) Record {
	return Record{
		Time:    r.Time,
		Message: r.Message,
		Level:   Level(r.Level),
	}
}

// EventFunc is a function to be executed when configured against a log level.
type EventFunc func(r Record)

// Events contains an assignment of an event function to a log level.
type Events struct {
	Debug EventFunc
	Info  EventFunc
	Warn  EventFunc
	Error EventFunc
}

// =============================================================================

// eventHandler provides a wrapper around the slog handler to capture which
// log level is being logged for event handling.
type eventHandler struct {
	handler slog.Handler
	events  Events
}

func newEventHandler(handler slog.Handler, events Events) *eventHandler {
	return &eventHandler{
		handler: handler,
		events:  events,
	}
}

// Enabled reports whether the handler handles records at the given level.
// The handler ignores records whose level is lower.
func (h *eventHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// WithAttrs returns a new JSONHandler whose attributes consists
// of h's attributes followed by attrs.
func (h *eventHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &eventHandler{handler: h.handler.WithAttrs(attrs), events: h.events}
}

// WithGroup returns a new Handler with the given group appended to the receiver's
// existing groups. The keys of all subsequent attributes, whether added by With
// or in a Record, should be qualified by the sequence of group names.
func (h *eventHandler) WithGroup(name string) slog.Handler {
	return &eventHandler{handler: h.handler.WithGroup(name), events: h.events}
}

// Handle looks to see if an event function needs to be executed for a given
// log level and then formats its argument Record.
func (h *eventHandler) Handle(ctx context.Context, r slog.Record) error {
	switch r.Level {
	case slog.LevelDebug:
		if h.events.Debug != nil {
			h.events.Debug(toRecord(r))
		}

	case slog.LevelError:
		if h.events.Error != nil {
			h.events.Error(toRecord(r))
		}

	case slog.LevelWarn:
		if h.events.Warn != nil {
			h.events.Warn(toRecord(r))
		}

	case slog.LevelInfo:
		if h.events.Info != nil {
			h.events.Info(toRecord(r))
		}
	}

	return h.handler.Handle(ctx, r)
}

// =============================================================================

// Logger represents a logger for logging information.
type Logger struct {
	*slog.Logger
}

// New constructs a new log for application use.
func New(w io.Writer, minLevel Level, serviceName string) *Logger {
	return new(w, minLevel, serviceName, nil)
}

// NewWithEvents constructs a new log for application use with events.
func NewWithEvents(w io.Writer, minLevel Level, serviceName string, events Events) *Logger {
	return new(w, minLevel, serviceName, &events)
}

// NewStdLogger returns a standard library Logger that wraps the slog Logger.
func NewStdLogger(logger *Logger, level Level) *log.Logger {
	return slog.NewLogLogger(logger.Handler(), slog.Level(level))
}

// Infoc logs the information at the specified call stack position.
func (log *Logger) Infoc(caller int, msg string, args ...any) {
	var pcs [1]uintptr
	runtime.Callers(caller, pcs[:]) // skip [Callers, Infof]

	r := slog.NewRecord(time.Now(), slog.LevelInfo, msg, pcs[0])
	r.Add(args...)

	log.Handler().Handle(context.Background(), r)
}

// =============================================================================

func new(w io.Writer, minLevel Level, serviceName string, events *Events) *Logger {

	// Convert the file name to just the name.ext when this key/value will
	// be logged.
	f := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			if source, ok := a.Value.Any().(*slog.Source); ok {
				v := fmt.Sprintf("%s:%d", filepath.Base(source.File), source.Line)
				return slog.Attr{Key: "file", Value: slog.StringValue(v)}
			}
		}

		return a
	}

	// Construct the slog JSON handler for use.
	handler := slog.Handler(slog.NewJSONHandler(w, &slog.HandlerOptions{AddSource: true, Level: slog.Level(minLevel), ReplaceAttr: f}))

	// If events are to be processed, wrap the JSON handler around the custom
	// event handler.
	if events != nil {
		handler = newEventHandler(handler, *events)
	}

	// Construct a logger.
	log := slog.New(handler)

	// Add the service name to the list of key/value pairs for each log line.
	if serviceName != "" {
		log = log.With("service", serviceName)
	}

	return &Logger{
		Logger: log,
	}
}
