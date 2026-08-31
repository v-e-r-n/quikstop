package otp

import (
	"context"
	"log"
)

// Deliverer defines the interface for sending verification codes to users.
type Deliverer interface {
	Deliver(ctx context.Context, recipient string, code string) error
}

// consoleDeliverer prints the code to standard output. Useful for development.
type consoleDeliverer struct {
	logger *log.Logger
}

// NewConsoleDeliverer creates a Deliverer that logs to stdout.
func NewConsoleDeliverer(logger *log.Logger) Deliverer {
	if logger == nil {
		logger = log.Default()
	}
	return &consoleDeliverer{logger: logger}
}

func (d *consoleDeliverer) Deliver(ctx context.Context, recipient string, code string) error {
	d.logger.Printf("[otp] Deliver code %s to %s", code, recipient)
	return nil
}
