package quikstop

import (
	"log/slog"

	"github.com/v-e-r-n/quikstop/core"
)

// SetLogger configures the global logger used across all quikstop subpackages.
func SetLogger(l *slog.Logger) {
	core.SetLogger(l)
}

// Logger returns the active global logger.
func Logger() *slog.Logger {
	return core.Logger()
}
