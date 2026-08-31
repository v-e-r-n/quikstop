package otp

import (
	"encoding/json"
	"errors"
	"net/http"
)

// SuccessCallback is triggered when a code is successfully verified.
// Use this to issue JWTs, establish session cookies, or log in users.
type SuccessCallback func(w http.ResponseWriter, r *http.Request, recipient string)

// Handler provides pre-built http.HandlerFunc wrappers for the OTP flows.
type Handler struct {
	keeper    *Keeper
	onSuccess SuccessCallback
}

// NewHandler creates a new HTTP Handler instance.
func NewHandler(keeper *Keeper, onSuccess SuccessCallback) *Handler {
	return &Handler{
		keeper:    keeper,
		onSuccess: onSuccess,
	}
}

// KnockRequest defines the incoming JSON payload for requesting a code.
type KnockRequest struct {
	Recipient string `json:"recipient"`
	Email     string `json:"email"` // fallback convenient key
	Phone     string `json:"phone"` // fallback convenient key
}

// VerifyRequest defines the incoming JSON payload for verifying a code.
type VerifyRequest struct {
	Recipient string `json:"recipient"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Code      string `json:"code"`
}

// Knock handles the request to generate and send an OTP code.
func (h *Handler) Knock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req KnockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	recipient := req.Recipient
	if recipient == "" {
		recipient = req.Email
	}
	if recipient == "" {
		recipient = req.Phone
	}

	if recipient == "" {
		h.writeError(w, "recipient is required", http.StatusBadRequest)
		return
	}

	if err := h.keeper.Knock(r.Context(), recipient); err != nil {
		h.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Verify checks the submitted code, handles errors, and executes the success callback.
func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	recipient := req.Recipient
	if recipient == "" {
		recipient = req.Email
	}
	if recipient == "" {
		recipient = req.Phone
	}

	if recipient == "" || req.Code == "" {
		h.writeError(w, "recipient and code are required", http.StatusBadRequest)
		return
	}

	ok, err := h.keeper.Verify(r.Context(), recipient, req.Code)
	if err != nil {
		if errors.Is(err, ErrInvalidCode) {
			h.writeError(w, "invalid_code", http.StatusUnauthorized)
			return
		}
		if errors.Is(err, ErrMaxRetriesExceeded) {
			h.writeError(w, "max_retries_exceeded", http.StatusForbidden)
			return
		}
		if errors.Is(err, ErrExpired) {
			h.writeError(w, "code_expired", http.StatusGone)
			return
		}
		if errors.Is(err, ErrNotFound) {
			h.writeError(w, "not_found", http.StatusNotFound)
			return
		}
		h.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !ok {
		h.writeError(w, "verification failed", http.StatusBadRequest)
		return
	}

	// Code was verified successfully! Trigger the callback.
	h.onSuccess(w, r, recipient)
}

func (h *Handler) writeError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
