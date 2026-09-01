package core

import (
	"log/slog"
	"sync/atomic"
)

var defaultLogger atomic.Pointer[slog.Logger]

func init() {
	defaultLogger.Store(slog.Default())
}

// SetLogger sets the global logger used across quikstop subpackages.
func SetLogger(l *slog.Logger) {
	if l == nil {
		l = slog.Default()
	}
	defaultLogger.Store(l)
}

// Logger returns the current active logger.
func Logger() *slog.Logger {
	return defaultLogger.Load()
}
