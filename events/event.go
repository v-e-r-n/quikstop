package events

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EventType is a typed string representing the type of event.
type EventType string

// Event represents a structured event notification payload.
type Event struct {
	ID        string       `json:"id"`
	Type      EventType    `json:"type"`
	Timestamp string       `json:"timestamp"`
	Payload   EventPayload `json:"payload"`
	Meta      EventMeta    `json:"meta"`
}

// EventPayload holds the resource-specific data for an event.
type EventPayload struct {
	ResourceType string      `json:"resource_type,omitempty"`
	ResourceID   string      `json:"resource_id,omitempty"`
	Data         interface{} `json:"data,omitempty"`
}

// EventMeta holds metadata about the event.
type EventMeta struct {
	ScopeID string                 `json:"scope_id,omitempty"`
	Extra   map[string]interface{} `json:"extra,omitempty"`
}

// NewEvent creates an event with a generated UUID and current UTC RFC3339 timestamp.
func NewEvent(eventType EventType, payload EventPayload, meta EventMeta) Event {
	return Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Payload:   payload,
		Meta:      meta,
	}
}

// ClientSession represents the authenticated state of a client connection.
type ClientSession struct {
	ClientID string                 `json:"client_id"`
	ScopeID  string                 `json:"scope_id,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Connection represents a single live client connection.
type Connection struct {
	ID      string
	Session ClientSession
	Channel chan []byte
	closed  bool
	mu      sync.RWMutex
}

// RecipientResolver determines which client IDs should receive a given event.
type RecipientResolver interface {
	ResolveClientIDs(ctx context.Context, ev Event) ([]string, error)
}
