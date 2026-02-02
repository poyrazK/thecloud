package mock

import "log/slog"

func NewNoopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(&NoopWriter{}, nil))
}

type NoopWriter struct{}

func (w *NoopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
