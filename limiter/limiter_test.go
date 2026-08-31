package limiter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/v-e-r-n/quikstop/limiter"
)

func TestLimiterAllow(t *testing.T) {
	// Rate: 10 tokens per second, Burst: 2
	lim := limiter.New(10, 2)

	ip := "1.2.3.4"

	// 1st request should be allowed (burst starts full)
	if !lim.Allow(ip) {
		t.Error("Expected 1st request to be allowed")
	}

	// 2nd request should be allowed (consuming the last token)
	if !lim.Allow(ip) {
		t.Error("Expected 2nd request to be allowed")
	}

	// 3rd request should be blocked (no tokens left)
	if lim.Allow(ip) {
		t.Error("Expected 3rd request to be blocked")
	}

	// Wait 150ms for tokens to refill (rate is 10/sec -> 1.5 tokens refilled)
	time.Sleep(150 * time.Millisecond)

	if !lim.Allow(ip) {
		t.Error("Expected request to be allowed after token refill")
	}
}

func TestMiddleware(t *testing.T) {
	var callbackCalled bool
	var callbackIP string

	lim := limiter.New(0.1, 1, limiter.WithOnLimit(func(ip string, r *http.Request) {
		callbackCalled = true
		callbackIP = ip
	}))

	mw := limiter.Handler(lim)

	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := mw(dummyHandler)

	// 1st request -> allowed
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "9.9.9.9:1234"
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got: %d", rr.Code)
	}

	// 2nd request -> blocked
	rr2 := httptest.NewRecorder()
	server.ServeHTTP(rr2, req)

	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("Expected 429 Too Many Requests, got: %d", rr2.Code)
	}

	if !callbackCalled {
		t.Error("Expected onLimit callback to be executed")
	}

	if callbackIP != "9.9.9.9" {
		t.Errorf("Expected callback IP to be 9.9.9.9, got: %s", callbackIP)
	}
}
