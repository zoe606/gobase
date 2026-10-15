package tasks_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/worker/tasks"
)

type authEmailProcessor struct {
	userID uint
	email  string
	err    error
}

func (p *authEmailProcessor) ProcessVerificationEmail(_ context.Context, id uint) error {
	p.userID = id
	return p.err
}

func (p *authEmailProcessor) ProcessPasswordResetEmail(_ context.Context, address string) error {
	p.email = address
	return p.err
}

func TestAuthEmailHandler(t *testing.T) {
	t.Parallel()
	failure := errors.New("sender unavailable")
	for _, tt := range []struct {
		name, kind, payload, email string
		userID                     uint
		processorErr, wantErr      error
	}{
		{name: "verification", kind: tasks.TypeEmailVerification, payload: `{"user_id":7}`, userID: 7},
		{name: "reset", kind: tasks.TypePasswordReset, payload: `{"email":"user@example.com"}`, email: "user@example.com"},
		{name: "delivery retry", kind: tasks.TypePasswordReset, payload: `{"email":"user@example.com"}`, email: "user@example.com", processorErr: failure, wantErr: failure},
		{name: "invalid JSON", kind: tasks.TypeEmailVerification, payload: `{`, wantErr: asynq.SkipRetry},
		{name: "missing user", kind: tasks.TypeEmailVerification, payload: `{}`, wantErr: asynq.SkipRetry},
		{name: "invalid email", kind: tasks.TypePasswordReset, payload: `{"email":"invalid"}`, wantErr: asynq.SkipRetry},
		{name: "display name", kind: tasks.TypePasswordReset, payload: `{"email":"Name <user@example.com>"}`, wantErr: asynq.SkipRetry},
		{name: "unknown task", kind: "auth:unknown", payload: `{}`, wantErr: asynq.SkipRetry},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			processor := &authEmailProcessor{err: tt.processorErr}
			err := tasks.AuthEmailHandler(processor).ProcessTask(t.Context(), asynq.NewTask(tt.kind, []byte(tt.payload)))
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.userID, processor.userID)
			require.Equal(t, tt.email, processor.email)
		})
	}
}
