package tasks

import (
	"context"
	"fmt"
	"net/mail"

	"github.com/hibiken/asynq"

	"go-boilerplate/pkg/json"
)

type AuthEmailPayload struct {
	UserID uint   `json:"user_id,omitempty"`
	Email  string `json:"email,omitempty"`
}

type AuthEmailProcessor interface {
	ProcessVerificationEmail(context.Context, uint) error
	ProcessPasswordResetEmail(context.Context, string) error
}

func AuthEmailHandler(processor AuthEmailProcessor) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload AuthEmailPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("decode auth email request: %w: %w", err, asynq.SkipRetry)
		}
		switch task.Type() {
		case TypeEmailVerification:
			if payload.UserID == 0 {
				return fmt.Errorf("missing verification user: %w", asynq.SkipRetry)
			}
			return processor.ProcessVerificationEmail(ctx, payload.UserID)
		case TypePasswordReset:
			address, err := mail.ParseAddress(payload.Email)
			if err != nil || address.Address != payload.Email || len(payload.Email) > 254 {
				return fmt.Errorf("invalid reset email: %w", asynq.SkipRetry)
			}
			return processor.ProcessPasswordResetEmail(ctx, payload.Email)
		default:
			return fmt.Errorf("unknown auth email request: %w", asynq.SkipRetry)
		}
	}
}
