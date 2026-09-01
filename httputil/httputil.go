package httputil

import (
	"encoding/json"
	"net/http"
)

// ErrorPayload represents a standardized JSON error response.
type ErrorPayload struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// JSON writes the given data as a JSON response with the provided status code.
func JSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data == nil {
		return nil
	}
	return json.NewEncoder(w).Encode(data)
}

// Error writes a standardized JSON error response with the provided status code and error code/message.
func Error(w http.ResponseWriter, status int, code string, message ...string) error {
	payload := ErrorPayload{Error: code}
	if len(message) > 0 && message[0] != "" {
		payload.Message = message[0]
	}
	return JSON(w, status, payload)
}

// DecodeJSON decodes the JSON request body into the target struct.
func DecodeJSON(r *http.Request, target any) error {
	return json.NewDecoder(r.Body).Decode(target)
}
