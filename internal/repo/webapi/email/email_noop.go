package email

import (
	"context"

	"go-boilerplate/pkg/logger"
)

// NoopSender is a no-operation email sender for development/testing.
// It logs emails instead of sending them.
type NoopSender struct {
	l logger.Interface
}

// NewNoopSender creates a new noop email sender.
func NewNoopSender(l logger.Interface) *NoopSender {
	return &NoopSender{l: l}
}

// Send logs the email instead of sending it.
func (n *NoopSender) Send(_ context.Context, em Email) error {
	n.l.Info("Noop mailer: would send email",
		"to", em.To,
		"subject", em.Subject,
	)

	return nil
}
