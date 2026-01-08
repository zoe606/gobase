package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"go-boilerplate/pkg/logger"
)

// EmailPayload represents the payload for email notification tasks.
type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// EmailHandler handles email notification tasks.
type EmailHandler struct {
	l logger.Interface
}

// NewEmailHandler creates a new email handler.
func NewEmailHandler(l logger.Interface) *EmailHandler {
	return &EmailHandler{l: l}
}

// ProcessTask processes an email notification task.
func (h *EmailHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload EmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal email payload: %w", err)
	}

	h.l.Info("Processing email notification",
		"to", payload.To,
		"subject", payload.Subject,
	)

	// TODO: Implement actual email sending logic here
	// For now, just log the email details

	h.l.Info("Email notification sent successfully",
		"to", payload.To,
	)

	return nil
}

// NewEmailTask creates a new email notification task.
func NewEmailTask(to, subject, body string) (*asynq.Task, error) {
	payload := EmailPayload{
		To:      to,
		Subject: subject,
		Body:    body,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeEmailNotification, data), nil
}
