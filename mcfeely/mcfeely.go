package mcfeely

import (
	"context"
	"log"
)

// McFeely handles speedy delivery of messages, particularly emails.
type McFeely interface {
	Send(ctx context.Context, to, subject, body string) error
}

// ConsoleMcFeely is a development implementation of McFeely that logs to stdout.
type ConsoleMcFeely struct{}

func NewConsoleMcFeely() McFeely {
	return &ConsoleMcFeely{}
}

func (c *ConsoleMcFeely) Send(ctx context.Context, to, subject, body string) error {
	log.Printf("\n--- [McFeely Speedy Delivery] ---\nTo: %s\nSubject: %s\nBody: %s\n---------------------------------\n", to, subject, body)
	return nil
}
