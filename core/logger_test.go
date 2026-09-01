package core_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/v-e-r-n/quikstop/core"
)

func TestCoreLogger(t *testing.T) {
	var buf bytes.Buffer
	customLogger := slog.New(slog.NewTextHandler(&buf, nil))

	core.SetLogger(customLogger)

	if core.Logger() != customLogger {
		t.Fatal("Expected core.Logger() to return customLogger")
	}

	core.Logger().Info("test message from core")

	if !strings.Contains(buf.String(), "test message from core") {
		t.Errorf("Expected buffer to contain 'test message from core', got: %s", buf.String())
	}
}
