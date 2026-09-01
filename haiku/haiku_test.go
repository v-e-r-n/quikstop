package haiku_test

import (
	"strings"
	"testing"

	"github.com/v-e-r-n/quikstop/haiku"
)

func TestGenerate(t *testing.T) {
	h1 := haiku.Generate()
	if h1 == "" {
		t.Fatal("expected non-empty haiku")
	}

	parts := strings.Split(h1, "-")
	if len(parts) != 3 {
		t.Fatalf("expected 3 hyphenated parts, got %d from %q", len(parts), h1)
	}

	if len(parts[2]) != 4 {
		t.Fatalf("expected 4-character hex suffix, got %q", parts[2])
	}

	// Verify randomness
	h2 := haiku.Generate()
	if h1 == h2 {
		t.Fatalf("expected different haikus, got duplicates: %q", h1)
	}
}
