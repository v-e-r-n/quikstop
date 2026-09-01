package jwt_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/v-e-r-n/quikstop/jwt"
)

var testSecret = []byte("super-secret-key-1234567890123456")

func TestGenerateAndVerify(t *testing.T) {
	token, err := jwt.Generate("user-123", testSecret, 1*time.Hour, "org-456")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	claims, err := jwt.Verify(token, testSecret)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if claims.UserID != "user-123" {
		t.Errorf("Expected UserID user-123, got: %s", claims.UserID)
	}
	if claims.ScopeID != "org-456" {
		t.Errorf("Expected ScopeID org-456, got: %s", claims.ScopeID)
	}
}

func TestVerifyExpired(t *testing.T) {
	token, err := jwt.Generate("user-123", testSecret, -1*time.Minute)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	_, err = jwt.Verify(token, testSecret)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestMiddleware(t *testing.T) {
	token, _ := jwt.Generate("user-999", testSecret, 15*time.Minute)
	mw := jwt.Middleware(testSecret)

	var extractedUserID, extractedScopeID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		extractedUserID = jwt.UserIDFromContext(r.Context())
		extractedScopeID = jwt.ScopeIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := mw(next)

	// 1. Authorization: Bearer <token> with X-Squad-Id header
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Squad-Id", "squad-alpha")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", rec.Code)
	}
	if extractedUserID != "user-999" {
		t.Errorf("Expected user-999, got: %s", extractedUserID)
	}
	if extractedScopeID != "squad-alpha" {
		t.Errorf("Expected squad-alpha, got: %s", extractedScopeID)
	}

	// 2. Query param ?token=<token> fallback
	req2 := httptest.NewRequest("GET", "/test?token="+token+"&squad_id=squad-beta", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", rec2.Code)
	}
	if extractedUserID != "user-999" {
		t.Errorf("Expected user-999, got: %s", extractedUserID)
	}
	if extractedScopeID != "squad-beta" {
		t.Errorf("Expected squad-beta, got: %s", extractedScopeID)
	}

	// 3. Unauthorized when token missing
	req3 := httptest.NewRequest("GET", "/test", nil)
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)

	if rec3.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 Unauthorized, got: %d", rec3.Code)
	}
}

func TestRefreshTokens(t *testing.T) {
	token, expiresAt, err := jwt.GenerateRefreshToken("user-refresh-123", testSecret, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}
	if expiresAt.Before(time.Now()) {
		t.Error("Expected future expiresAt")
	}

	userID, err := jwt.VerifyRefreshToken(token, testSecret)
	if err != nil {
		t.Fatalf("VerifyRefreshToken failed: %v", err)
	}
	if userID != "user-refresh-123" {
		t.Errorf("Expected user-refresh-123, got: %s", userID)
	}

	// Access token should fail refresh token verification
	accessToken, _ := jwt.Generate("user-refresh-123", testSecret, 1*time.Hour)
	_, err = jwt.VerifyRefreshToken(accessToken, testSecret)
	if err == nil {
		t.Error("Expected error when verifying access token as refresh token")
	}
}

