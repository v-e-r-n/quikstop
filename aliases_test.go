package quikstop_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/v-e-r-n/quikstop"
)

func TestRootAliases(t *testing.T) {
	// 1. JWT Aliases
	secret := []byte("test-secret-key-1234567890123456")
	token, err := quikstop.GenerateJWT("user-1", secret, 1*time.Hour, "org-1")
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}

	claims, err := quikstop.VerifyJWT(token, secret)
	if err != nil || claims.UserID != "user-1" {
		t.Fatalf("VerifyJWT failed: %v", err)
	}

	// 2. CORS & HTTP Utilities
	mw := quikstop.CORS()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected CORS preflight 200, got: %d", rec.Code)
	}

	recJSON := httptest.NewRecorder()
	quikstop.JSON(recJSON, http.StatusAccepted, map[string]string{"status": "ok"})
	if recJSON.Code != http.StatusAccepted {
		t.Errorf("Expected JSON status 202, got: %d", recJSON.Code)
	}

	// 3. Limiter Aliases
	limiter := quikstop.NewLimiter(10, 10)
	if !limiter.Allow("127.0.0.1") {
		t.Error("Expected limiter to allow request")
	}

	// 4. Events Aliases
	bus := quikstop.NewEventBus()
	ch := bus.Subscribe()
	defer bus.Unsubscribe(ch)
	bus.Publish(quikstop.NewEvent("TEST", quikstop.EventPayload{}, quikstop.EventMeta{}))
	select {
	case <-ch:
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected event to be received via root alias")
	}

	// 5. OTP Aliases
	store := quikstop.NewMemoryOTPStore()
	deliverer := quikstop.NewConsoleDeliverer()
	keeper := quikstop.NewOTP(store, deliverer, quikstop.OTPConfig{
		CodeLength: 6,
		TTL:        1 * time.Minute,
		MaxRetries: 3,
	})
	_ = keeper.Knock(context.Background(), "test@example.com")

	// 6. McFeely Aliases
	mailer := quikstop.NewConsoleMcFeely()
	if err := mailer.Send(context.Background(), "user@example.com", "Subject", "Body"); err != nil {
		t.Errorf("Expected nil error from NewConsoleMcFeely, got: %v", err)
	}
}
