package logger

import "io"

// Discard is a no-op writer that can be used to discard log output.
var Discard io.Writer = noopWriter{}

type noopWriter struct{}

func (noopWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func (noopWriter) WriteString(s string) (int, error) {
	return len(s), nil
}
