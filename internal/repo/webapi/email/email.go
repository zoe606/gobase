// Package email provides email service integrations.
package email

import "context"

// Sender defines the interface for sending emails.
// Implementations: Resend, Noop (for development).
type Sender interface {
	Send(ctx context.Context, email Email) error
}

// Email represents an email to be sent.
type Email struct {
	To      []string
	From    string
	Subject string
	HTML    string
	Text    string
}
