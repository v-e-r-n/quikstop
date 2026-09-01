package otp

import (
	"context"

	"github.com/v-e-r-n/quikstop/core"
)

// Deliverer defines the interface for sending verification codes to users.
type Deliverer interface {
	Deliver(ctx context.Context, recipient string, code string) error
}

// consoleDeliverer prints the code using the active logger. Useful for development.
type consoleDeliverer struct{}

// NewConsoleDeliverer creates a Deliverer that logs using the core active logger.
func NewConsoleDeliverer() Deliverer {
	return &consoleDeliverer{}
}

func (d *consoleDeliverer) Deliver(ctx context.Context, recipient string, code string) error {
	core.Logger().Info("Deliver verification code",
		"component", "otp",
		"recipient", recipient,
		"code", code,
	)
	return nil
}
