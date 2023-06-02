// Package logger provides support for initializing the log system.
package logger

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/exp/slog"
)

// Logger represents a logger for logging information.
type Logger struct {
	*slog.Logger
}

// New constructs a new log for application use.
func New(w io.Writer, serviceName string) *Logger {

	// Convert the file and line number to a single attribute.
	f := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			if source, ok := a.Value.Any().(*slog.Source); ok {
				v := fmt.Sprintf("%s:%d", filepath.Base(source.File), source.Line)
				return slog.Attr{Key: "file", Value: slog.StringValue(v)}
			}
		}

		return a
	}

	log := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{AddSource: true, ReplaceAttr: f}))

	if serviceName != "" {
		log = log.With("service", serviceName)
	}

	return &Logger{
		Logger: log,
	}
}

// Infoc logs the information at the specified call stack position.
func (log *Logger) Infoc(caller int, msg string, args ...any) {
	var pcs [1]uintptr
	runtime.Callers(caller, pcs[:]) // skip [Callers, Infof]

	r := slog.NewRecord(time.Now(), slog.LevelInfo, msg, pcs[0])
	r.Add(args...)

	log.Handler().Handle(context.Background(), r)
}
