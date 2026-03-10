package tasks_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"go-boilerplate/internal/repo/webapi/email"
	"go-boilerplate/internal/worker/tasks"
	"go-boilerplate/pkg/json"
)

type mockLogger struct{}

func (m *mockLogger) Debug(_ interface{}, _ ...interface{}) {}
func (m *mockLogger) Info(_ string, _ ...interface{})       {}
func (m *mockLogger) Warn(_ string, _ ...interface{})       {}
func (m *mockLogger) Error(_ interface{}, _ ...interface{}) {}
func (m *mockLogger) Fatal(_ interface{}, _ ...interface{}) {}
func (m *mockLogger) GetZapLogger() *zap.Logger             { return nil }

type mockMailer struct {
	sendFn func(ctx context.Context, e email.Email) error
}

func (m *mockMailer) Send(ctx context.Context, e email.Email) error {
	if m.sendFn != nil {
		return m.sendFn(ctx, e)
	}
	return nil
}

func TestEmailHandler_ProcessTask_Success(t *testing.T) {
	t.Parallel()

	mailer := &mockMailer{}
	handler := tasks.NewEmailHandler(&mockLogger{}, mailer)

	payload := tasks.EmailPayload{
		To:       "test@example.com",
		Subject:  "Test",
		HTML:     "<p>Hello</p>",
		Text:     "Hello",
		TaskType: "generic",
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(tasks.TypeEmailNotification, data)
	err = handler.ProcessTask(context.Background(), task)
	require.NoError(t, err)
}

func TestEmailHandler_ProcessTask_BadPayload(t *testing.T) {
	t.Parallel()

	handler := tasks.NewEmailHandler(&mockLogger{}, &mockMailer{})

	task := asynq.NewTask(tasks.TypeEmailNotification, []byte(`{invalid`))
	err := handler.ProcessTask(context.Background(), task)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unmarshal email payload")
}

func TestEmailHandler_ProcessTask_SendError(t *testing.T) {
	t.Parallel()

	sendErr := errors.New("smtp failure")
	mailer := &mockMailer{
		sendFn: func(_ context.Context, _ email.Email) error {
			return sendErr
		},
	}

	handler := tasks.NewEmailHandler(&mockLogger{}, mailer)

	payload := tasks.EmailPayload{
		To:       "test@example.com",
		Subject:  "Test",
		HTML:     "<p>Hello</p>",
		Text:     "Hello",
		TaskType: "generic",
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(tasks.TypeEmailNotification, data)
	err = handler.ProcessTask(context.Background(), task)
	require.Error(t, err)
	require.Contains(t, err.Error(), "send email")
}

func TestNewEmailTask(t *testing.T) {
	t.Parallel()

	task, err := tasks.NewEmailTask("to@test.com", "Subject", "<p>HTML</p>", "Text")
	require.NoError(t, err)
	require.Equal(t, tasks.TypeEmailNotification, task.Type())
	require.NotEmpty(t, task.Payload())
}

func TestNewWelcomeEmailTask(t *testing.T) {
	t.Parallel()

	task, err := tasks.NewWelcomeEmailTask("user@test.com", "Alice", "MyApp")
	require.NoError(t, err)
	require.Equal(t, tasks.TypeEmailNotification, task.Type())

	var payload tasks.EmailPayload
	err = json.Unmarshal(task.Payload(), &payload)
	require.NoError(t, err)
	require.Equal(t, "user@test.com", payload.To)
	require.Equal(t, "welcome", payload.TaskType)
	require.Contains(t, payload.Subject, "MyApp")
}
