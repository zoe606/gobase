package tasks

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"

	internalemail "go-boilerplate/internal/email"
	"go-boilerplate/internal/repo/webapi/email"
	"go-boilerplate/pkg/json"
	"go-boilerplate/pkg/logger"
)

// EmailPayload represents the payload for email notification tasks.
type EmailPayload struct {
	To       string `json:"to"`
	Subject  string `json:"subject"`
	HTML     string `json:"html"`
	Text     string `json:"text"`
	TaskType string `json:"task_type"` // welcome, reset_password, etc.
}

// EmailHandler handles email notification tasks.
type EmailHandler struct {
	l      logger.Interface
	mailer email.Sender
}

// NewEmailHandler creates a new email handler.
func NewEmailHandler(l logger.Interface, mailer email.Sender) *EmailHandler {
	return &EmailHandler{
		l:      l,
		mailer: mailer,
	}
}

// ProcessTask processes an email notification task.
func (h *EmailHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload EmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal email payload: %w", err)
	}

	h.l.Info("Sending email",
		"to", payload.To,
		"subject", payload.Subject,
		"type", payload.TaskType,
	)

	err := h.mailer.Send(ctx, email.Email{
		To:      []string{payload.To},
		Subject: payload.Subject,
		HTML:    payload.HTML,
		Text:    payload.Text,
	})
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	h.l.Info("Email sent",
		"to", payload.To,
		"type", payload.TaskType,
	)

	return nil
}

// NewEmailTask creates a new email notification task.
func NewEmailTask(to, subject, html, text string) (*asynq.Task, error) {
	payload := EmailPayload{
		To:       to,
		Subject:  subject,
		HTML:     html,
		Text:     text,
		TaskType: "generic",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeEmailNotification, data), nil
}

// NewWelcomeEmailTask creates a welcome email task.
func NewWelcomeEmailTask(to, username, appName string) (*asynq.Task, error) {
	html, text, err := internalemail.RenderWelcomeEmail(internalemail.WelcomeEmailData{
		Username: username,
		AppName:  appName,
	})
	if err != nil {
		return nil, fmt.Errorf("render welcome email: %w", err)
	}

	payload := EmailPayload{
		To:       to,
		Subject:  "Welcome to " + appName,
		HTML:     html,
		Text:     text,
		TaskType: "welcome",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeEmailNotification, data), nil
}
