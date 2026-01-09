package email

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v2"
)

// ResendSender sends emails using the Resend API.
type ResendSender struct {
	client *resend.Client
	from   string
}

// NewResendSender creates a new Resend email sender.
func NewResendSender(apiKey, fromEmail string) *ResendSender {
	return &ResendSender{
		client: resend.NewClient(apiKey),
		from:   fromEmail,
	}
}

// Send sends an email using Resend.
func (r *ResendSender) Send(ctx context.Context, email Email) error {
	params := &resend.SendEmailRequest{
		From:    r.from,
		To:      email.To,
		Subject: email.Subject,
		Html:    email.HTML,
		Text:    email.Text,
	}

	_, err := r.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("resend: %w", err)
	}

	return nil
}
