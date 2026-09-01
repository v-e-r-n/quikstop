package mcfeely

import (
	"context"

	"github.com/v-e-r-n/quikstop/core"
)

// McFeely handles speedy delivery of messages, particularly emails.
type McFeely interface {
	Send(ctx context.Context, to, subject, body string) error
}

// ConsoleMcFeely is a development implementation of McFeely that logs to the active logger.
type ConsoleMcFeely struct{}

func NewConsoleMcFeely() McFeely {
	return &ConsoleMcFeely{}
}

func (c *ConsoleMcFeely) Send(ctx context.Context, to, subject, body string) error {
	core.Logger().Info("Speedy delivery",
		"component", "mcfeely",
		"to", to,
		"subject", subject,
		"body", body,
	)
	return nil
}
