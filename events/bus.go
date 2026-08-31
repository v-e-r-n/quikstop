package events

import (
	"sync"
)

// EventBusOption allows configuring the EventBus instance.
type EventBusOption func(*EventBus)

// WithOnDrop registers a callback executed when an event is dropped due to channel capacity limits.
func WithOnDrop(cb func(ev Event)) EventBusOption {
	return func(eb *EventBus) {
		eb.onDrop = cb
	}
}

// EventBus is the central pub/sub mechanism for real-time events.
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[chan Event]bool
	onDrop      func(ev Event)
}

// NewEventBus creates a new EventBus.
func NewEventBus(opts ...EventBusOption) *EventBus {
	eb := &EventBus{
		subscribers: make(map[chan Event]bool),
	}

	for _, opt := range opts {
		opt(eb)
	}

	return eb
}

// Subscribe registers a new subscriber channel.
func (eb *EventBus) Subscribe() chan Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	ch := make(chan Event, 100)
	eb.subscribers[ch] = true
	return ch
}

// Unsubscribe removes a subscriber channel.
func (eb *EventBus) Unsubscribe(ch chan Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if _, ok := eb.subscribers[ch]; ok {
		delete(eb.subscribers, ch)
		close(ch)
	}
}

// Publish distributes an event to all active subscribers in a non-blocking way.
func (eb *EventBus) Publish(ev Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	for ch := range eb.subscribers {
		select {
		case ch <- ev:
		default:
			if eb.onDrop != nil {
				eb.onDrop(ev)
			}
		}
	}
}

// SubscriberCount returns the current number of active subscribers.
func (eb *EventBus) SubscriberCount() int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return len(eb.subscribers)
}
