package httputil_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/v-e-r-n/quikstop/httputil"
)

func TestJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]string{"status": "ok"}

	err := httputil.JSON(rec, http.StatusCreated, data)
	if err != nil {
		t.Fatalf("JSON returned error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status 201 Created, got: %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected application/json header, got: %s", rec.Header().Get("Content-Type"))
	}

	var parsed map[string]string
	json.NewDecoder(rec.Body).Decode(&parsed)
	if parsed["status"] != "ok" {
		t.Errorf("Expected status: ok, got: %s", parsed["status"])
	}
}

func TestError(t *testing.T) {
	rec := httptest.NewRecorder()
	err := httputil.Error(rec, http.StatusBadRequest, "invalid_input", "The field is required")
	if err != nil {
		t.Fatalf("Error helper returned error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request, got: %d", rec.Code)
	}

	var parsed httputil.ErrorPayload
	json.NewDecoder(rec.Body).Decode(&parsed)
	if parsed.Error != "invalid_input" || parsed.Message != "The field is required" {
		t.Errorf("Unexpected error response: %+v", parsed)
	}
}

func TestDecodeJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"test-user"}`))
	var body struct {
		Name string `json:"name"`
	}

	err := httputil.DecodeJSON(req, &body)
	if err != nil {
		t.Fatalf("DecodeJSON failed: %v", err)
	}
	if body.Name != "test-user" {
		t.Errorf("Expected name: test-user, got: %s", body.Name)
	}
}
