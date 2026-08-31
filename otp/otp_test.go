package otp_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/v-e-r-n/quikstop/otp"
)

type mockDeliverer struct {
	lastRecipient string
	lastCode      string
}

func (d *mockDeliverer) Deliver(ctx context.Context, recipient, code string) error {
	d.lastRecipient = recipient
	d.lastCode = code
	return nil
}

func TestKeeperFlow(t *testing.T) {
	store := otp.NewMemoryStore()
	deliverer := &mockDeliverer{}
	keeper := otp.New(store, deliverer, otp.Config{
		CodeLength: 6,
		TTL:        100 * time.Millisecond,
		MaxRetries: 3,
	})

	recipient := "test@example.com"
	ctx := context.Background()

	// 1. Send OTP
	if err := keeper.Knock(ctx, recipient); err != nil {
		t.Fatalf("Knock failed: %v", err)
	}

	code := deliverer.lastCode
	if len(code) != 6 {
		t.Errorf("Expected 6-digit code, got length: %d", len(code))
	}

	// 2. Verify wrong code
	ok, err := keeper.Verify(ctx, recipient, "wrong")
	if ok || !errors.Is(err, otp.ErrInvalidCode) {
		t.Errorf("Expected ErrInvalidCode, got: %v (ok: %t)", err, ok)
	}

	// 3. Verify correct code
	ok, err = keeper.Verify(ctx, recipient, code)
	if !ok || err != nil {
		t.Fatalf("Expected verification success, got: %v (ok: %t)", err, ok)
	}

	// 4. Verify code is deleted after success (returns NotFound)
	ok, err = keeper.Verify(ctx, recipient, code)
	if ok || !errors.Is(err, otp.ErrNotFound) {
		t.Errorf("Expected ErrNotFound after success deletion, got: %v", err)
	}
}

func TestKeeperExpiration(t *testing.T) {
	store := otp.NewMemoryStore()
	deliverer := &mockDeliverer{}
	keeper := otp.New(store, deliverer, otp.Config{
		CodeLength: 6,
		TTL:        10 * time.Millisecond,
		MaxRetries: 3,
	})

	recipient := "test@example.com"
	ctx := context.Background()

	_ = keeper.Knock(ctx, recipient)
	code := deliverer.lastCode

	time.Sleep(15 * time.Millisecond)

	ok, err := keeper.Verify(ctx, recipient, code)
	if ok || !errors.Is(err, otp.ErrExpired) {
		t.Errorf("Expected ErrExpired, got: %v", err)
	}
}

func TestKeeperMaxRetries(t *testing.T) {
	store := otp.NewMemoryStore()
	deliverer := &mockDeliverer{}
	keeper := otp.New(store, deliverer, otp.Config{
		CodeLength: 6,
		TTL:        5 * time.Minute,
		MaxRetries: 3,
	})

	recipient := "test@example.com"
	ctx := context.Background()

	_ = keeper.Knock(ctx, recipient)

	// 1st wrong attempt
	_, _ = keeper.Verify(ctx, recipient, "wrong1")
	// 2nd wrong attempt
	_, _ = keeper.Verify(ctx, recipient, "wrong2")
	// 3rd wrong attempt -> locks out (MaxRetriesExceeded)
	ok, err := keeper.Verify(ctx, recipient, "wrong3")
	if ok || !errors.Is(err, otp.ErrMaxRetriesExceeded) {
		t.Errorf("Expected ErrMaxRetriesExceeded, got: %v", err)
	}
}

func TestHTTPHandler(t *testing.T) {
	store := otp.NewMemoryStore()
	deliverer := &mockDeliverer{}
	keeper := otp.New(store, deliverer, otp.Config{
		CodeLength: 6,
		TTL:        5 * time.Minute,
		MaxRetries: 3,
	})

	var callbackCalled bool
	onSuccess := func(w http.ResponseWriter, r *http.Request, recipient string) {
		callbackCalled = true
		w.WriteHeader(http.StatusOK)
	}

	h := otp.NewHandler(keeper, onSuccess)

	// 1. Knock request
	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest("POST", "/otp", strings.NewReader(`{"recipient":"api@example.com"}`))
	h.Knock(w1, r1)

	if w1.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got: %d", w1.Code)
	}

	code := deliverer.lastCode

	// 2. Verify request (wrong code -> 401 Unauthorized)
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("POST", "/verify", strings.NewReader(`{"recipient":"api@example.com","code":"wrong"}`))
	h.Verify(w2, r2)

	if w2.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for wrong code, got: %d", w2.Code)
	}

	// 3. Verify request (correct code -> 200 OK)
	w3 := httptest.NewRecorder()
	r3 := httptest.NewRequest("POST", "/verify", strings.NewReader(`{"recipient":"api@example.com","code":"`+code+`"}`))
	h.Verify(w3, r3)

	if w3.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got: %d", w3.Code)
	}

	if !callbackCalled {
		t.Error("Expected onSuccess callback to be executed")
	}

	// 4. Decode body error response
	var errResp map[string]string
	json.NewDecoder(w2.Body).Decode(&errResp)
	if errResp["error"] != "invalid_code" {
		t.Errorf("Expected 'invalid_code' error response payload, got: %s", errResp["error"])
	}
}
