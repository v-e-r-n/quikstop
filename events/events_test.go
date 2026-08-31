package events_test

import (
	"context"
	"testing"
	"time"

	"github.com/v-e-r-n/quikstop/events"
)

type mockResolver struct{}

func (m *mockResolver) ResolveClientIDs(ctx context.Context, ev events.Event) ([]string, error) {
	if ev.Meta.ScopeID == "test-scope" {
		return []string{"user-1"}, nil
	}
	return nil, nil
}

func TestEventsEcosystem(t *testing.T) {
	bus := events.NewEventBus()
	sm := events.NewStreamManager()
	resolver := &mockResolver{}

	dispatcher := events.NewDispatcher(bus, sm, resolver)
	dispatcher.Start(context.Background())
	defer dispatcher.Stop()

	session := events.ClientSession{
		ClientID: "user-1",
		ScopeID:  "test-scope",
	}

	conn := sm.Connect(session)
	defer sm.Disconnect(conn.ID)

	event := events.NewEvent(
		events.EventType("TEST_EVENT"),
		events.EventPayload{Data: "hello"},
		events.EventMeta{ScopeID: "test-scope"},
	)

	bus.Publish(event)

	select {
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for event distribution")
	case data := <-conn.Channel:
		if !stringsContains(string(data), "hello") {
			t.Errorf("Expected event payload to contain 'hello', got: %s", string(data))
		}
	}
}

func stringsContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
