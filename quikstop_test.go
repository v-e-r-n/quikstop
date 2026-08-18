package quikstop

import (
	"os"
	"testing"
)

func TestEnvString(t *testing.T) {
	// 1. Fallback to default when unset
	os.Unsetenv("TEST_ENV_VAR")
	if val := EnvString("TEST_ENV_VAR", "default"); val != "default" {
		t.Errorf("Expected 'default', got %s", val)
	}

	// 2. Read value when set
	os.Setenv("TEST_ENV_VAR", "custom")
	defer os.Unsetenv("TEST_ENV_VAR")
	if val := EnvString("TEST_ENV_VAR", "default"); val != "custom" {
		t.Errorf("Expected 'custom', got %s", val)
	}
}

func TestResolveAddr(t *testing.T) {
	// Setup env
	os.Setenv("MOCK_HOST", "127.0.0.1")
	os.Setenv("MOCK_PORT", "9090")
	os.Setenv("GENERIC_PORT", "8080")
	defer func() {
		os.Unsetenv("MOCK_HOST")
		os.Unsetenv("MOCK_PORT")
		os.Unsetenv("GENERIC_PORT")
	}()

	// 1. Resolve host and specific port
	addr1 := ResolveAddr("MOCK_HOST", []string{"MOCK_PORT", "GENERIC_PORT"}, "8081")
	if addr1 != "127.0.0.1:9090" {
		t.Errorf("Expected '127.0.0.1:9090', got %s", addr1)
	}

	// 2. Resolve host and secondary port (precedence)
	os.Unsetenv("MOCK_PORT")
	addr2 := ResolveAddr("MOCK_HOST", []string{"MOCK_PORT", "GENERIC_PORT"}, "8081")
	if addr2 != "127.0.0.1:8080" {
		t.Errorf("Expected '127.0.0.1:8080', got %s", addr2)
	}

	// 3. Fallback port
	os.Unsetenv("GENERIC_PORT")
	addr3 := ResolveAddr("MOCK_HOST", []string{"MOCK_PORT", "GENERIC_PORT"}, "8081")
	if addr3 != "127.0.0.1:8081" {
		t.Errorf("Expected '127.0.0.1:8081', got %s", addr3)
	}

	// 4. Empty host
	os.Unsetenv("MOCK_HOST")
	addr4 := ResolveAddr("MOCK_HOST", []string{"MOCK_PORT", "GENERIC_PORT"}, "8081")
	if addr4 != ":8081" {
		t.Errorf("Expected ':8081', got %s", addr4)
	}
}
