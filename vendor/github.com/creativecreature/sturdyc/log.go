package sturdyc

type Logger interface {
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type NoopLogger struct{}

func (l *NoopLogger) Warn(_ string, _ ...any)  {}
func (l *NoopLogger) Error(_ string, _ ...any) {}
