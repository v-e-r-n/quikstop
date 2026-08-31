package events

import (
	"fmt"
	"net/http"
	"time"
)

// SSEConfig holds configuration settings for the Server-Sent Events handler.
type SSEConfig struct {
	HeartbeatInterval time.Duration
	// ContextExtractor verifies the request context and returns ClientSession authentication details.
	ContextExtractor  func(r *http.Request) (ClientSession, error)
}

// NewSSEHandler returns an http.HandlerFunc that handles SSE connections.
func NewSSEHandler(sm *StreamManager, cfg SSEConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := cfg.ContextExtractor(r)
		if err != nil || session.ClientID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Set headers required for Server-Sent Events
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no") // Disable buffering for Nginx if proxying

		conn := sm.Connect(session)
		defer sm.Disconnect(conn.ID)

		// Get HTTP flusher to stream data immediately
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		// Send initial open event
		fmt.Fprintf(w, "event: open\ndata: {\"status\":\"connected\",\"conn_id\":\"%s\"}\n\n", conn.ID)
		flusher.Flush()

		heartbeatTicker := time.NewTicker(cfg.HeartbeatInterval)
		defer heartbeatTicker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-heartbeatTicker.C:
				fmt.Fprint(w, ": ping\n\n")
				flusher.Flush()
			case data, ok := <-conn.Channel:
				if !ok {
					return
				}
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			}
		}
	}
}
