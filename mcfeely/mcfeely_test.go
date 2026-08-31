package mcfeely_test

import (
	"context"
	"testing"

	"github.com/v-e-r-n/quikstop/mcfeely"
)

func TestConsoleMcFeely(t *testing.T) {
	m := mcfeely.NewConsoleMcFeely()
	err := m.Send(context.Background(), "test@example.com", "Hello", "World")
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}
}
