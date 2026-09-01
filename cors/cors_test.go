package cors_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/v-e-r-n/quikstop/cors"
)

func TestCORSPreflight(t *testing.T) {
	mw := cors.Handler()
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	handler := mw(next)

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK for OPTIONS preflight, got: %d", rec.Code)
	}

	if called {
		t.Error("Expected preflight OPTIONS request to terminate early without calling next")
	}

	origin := rec.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin: *, got: %s", origin)
	}
}

func TestCORSPassthrough(t *testing.T) {
	mw := cors.Handler(cors.WithOrigins("https://example.com"))
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	})

	handler := mw(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("Expected next handler to be called on GET request")
	}
	if rec.Code != http.StatusTeapot {
		t.Errorf("Expected status 418, got: %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("Expected custom origin header, got: %s", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}
