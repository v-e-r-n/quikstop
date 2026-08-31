package events

import (
	"context"
	"encoding/json"
	"sync"
)

// Dispatcher coordinates event fanning out to resolved active clients.
type Dispatcher struct {
	bus      *EventBus
	sm       *StreamManager
	resolver RecipientResolver
	subCh    chan Event
	stopCh   chan struct{}
	wg       sync.WaitGroup
	once     sync.Once
}

// NewDispatcher creates a new Dispatcher.
func NewDispatcher(bus *EventBus, sm *StreamManager, resolver RecipientResolver) *Dispatcher {
	return &Dispatcher{
		bus:      bus,
		sm:       sm,
		resolver: resolver,
		subCh:    bus.Subscribe(),
		stopCh:   make(chan struct{}),
	}
}

// Start spawns the event dispatch loop.
func (d *Dispatcher) Start(ctx context.Context) {
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		defer d.bus.Unsubscribe(d.subCh)

		for {
			select {
			case <-d.stopCh:
				return
			case ev := <-d.subCh:
				// 1. Resolve targeted clients
				clientIDs, err := d.resolver.ResolveClientIDs(ctx, ev)
				if err != nil {
					continue
				}

				if len(clientIDs) == 0 {
					continue
				}

				// 2. Marshal to JSON once for all recipients
				data, err := json.Marshal(ev)
				if err != nil {
					continue
				}

				// 3. Fan out to active connections of resolved clients
				for _, clientID := range clientIDs {
					d.sm.BroadcastToClient(clientID, data)
				}
			}
		}
	}()
}

// Stop gracefully shuts down the dispatcher loop.
func (d *Dispatcher) Stop() {
	d.once.Do(func() {
		close(d.stopCh)
		d.wg.Wait()
	})
}
