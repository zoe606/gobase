package email

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockLogger implements logger.Interface for testing.
type mockLogger struct{}

func (m *mockLogger) Debug(msg interface{}, args ...interface{}) {}
func (m *mockLogger) Info(msg string, args ...interface{})       {}
func (m *mockLogger) Warn(msg string, args ...interface{})       {}
func (m *mockLogger) Error(msg interface{}, args ...interface{}) {}
func (m *mockLogger) Fatal(msg interface{}, args ...interface{}) {}
func (m *mockLogger) GetZapLogger() *zap.Logger                  { return zap.NewNop() }

func TestNoopSender_Send(t *testing.T) {
	l := &mockLogger{}
	sender := NewNoopSender(l)

	require.NotNil(t, sender)

	err := sender.Send(context.Background(), Email{
		To:      []string{"test@example.com"},
		Subject: "Test Subject",
		HTML:    "<p>Test body</p>",
		Text:    "Test body",
	})

	assert.NoError(t, err)
}

func TestResendSender_Creation(t *testing.T) {
	sender := NewResendSender("test-api-key", "noreply@example.com")

	require.NotNil(t, sender)
	assert.Equal(t, "noreply@example.com", sender.from)
	assert.NotNil(t, sender.client)
}

func TestEmail_Struct(t *testing.T) {
	email := Email{
		To:      []string{"user@example.com", "user2@example.com"},
		Subject: "Welcome",
		HTML:    "<h1>Hello</h1>",
		Text:    "Hello",
	}

	assert.Len(t, email.To, 2)
	assert.Equal(t, "Welcome", email.Subject)
	assert.Equal(t, "<h1>Hello</h1>", email.HTML)
	assert.Equal(t, "Hello", email.Text)
}

// TestSenderInterface verifies that both implementations satisfy the interface.
func TestSenderInterface(t *testing.T) {
	var _ Sender = (*NoopSender)(nil)
	var _ Sender = (*ResendSender)(nil)
}
