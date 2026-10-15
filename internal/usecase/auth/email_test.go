package auth_test

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/repo/webapi/email"
	"go-boilerplate/internal/usecase/auth"
	"go-boilerplate/pkg/audit"
)

type capturedAuthEmail struct {
	messages []email.Email
	err      error
}

func (s *capturedAuthEmail) Send(_ context.Context, message email.Email) error {
	s.messages = append(s.messages, message)
	return s.err
}

type authEmailFixture struct {
	uc     *auth.UseCase
	users  *MockUserRepo
	verify *MockEmailVerificationRepo
	resets *MockPasswordResetRepo
	sender *capturedAuthEmail
}

func newAuthEmailFixture(t *testing.T) authEmailFixture {
	t.Helper()
	ctrl := gomock.NewController(t)
	f := authEmailFixture{
		users: NewMockUserRepo(ctrl), verify: NewMockEmailVerificationRepo(ctrl),
		resets: NewMockPasswordResetRepo(ctrl), sender: &capturedAuthEmail{},
	}
	f.uc = auth.New(f.users, nil, nil, nil, audit.NewNoop()).
		WithAsynq(nil, "Example App").
		WithEmailVerification(f.verify, auth.VerificationConfig{Enabled: true, TokenTTL: time.Hour, BaseURL: "https://app.example/base"}).
		WithPasswordReset(f.resets, auth.ResetConfig{TokenTTL: time.Hour, BaseURL: "https://app.example/base"}).
		WithEmailDelivery(true, true).WithEmailSender(f.sender)
	return f
}

func (f authEmailFixture) expectPersistence(purpose string, deleteErr, createErr error) {
	if purpose == "verification" {
		f.verify.EXPECT().DeleteByUserID(gomock.Any(), uint(7)).Return(deleteErr)
		if deleteErr == nil {
			f.verify.EXPECT().Create(gomock.Any(), gomock.Any()).Return(createErr)
		}
		return
	}
	f.resets.EXPECT().DeleteByUserID(gomock.Any(), uint(7)).Return(deleteErr)
	if deleteErr == nil {
		f.resets.EXPECT().Create(gomock.Any(), gomock.Any()).Return(createErr)
	}
}

func TestAuthEmailProcessing(t *testing.T) {
	t.Parallel()
	for _, verification := range []bool{false, true} {
		name := "reset"
		if verification {
			name = "verification"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			f := newAuthEmailFixture(t)
			user := &entity.User{ID: 7, Email: "user@example.com", Active: true}
			var token string
			if verification {
				f.users.EXPECT().GetByID(gomock.Any(), uint(7)).Return(user, nil)
				f.verify.EXPECT().DeleteByUserID(gomock.Any(), uint(7)).Return(nil)
				f.verify.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, record *entity.EmailVerification) error {
					token = record.Token
					require.WithinDuration(t, time.Now().Add(time.Hour), record.ExpiresAt, time.Minute)
					return nil
				})
				require.NoError(t, f.uc.ProcessVerificationEmail(t.Context(), 7))
			} else {
				f.users.EXPECT().GetByEmail(gomock.Any(), user.Email).Return(user, nil)
				f.resets.EXPECT().DeleteByUserID(gomock.Any(), uint(7)).Return(nil)
				f.resets.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, record *entity.PasswordReset) error {
					token = record.Token
					require.WithinDuration(t, time.Now().Add(time.Hour), record.ExpiresAt, time.Minute)
					return nil
				})
				require.NoError(t, f.uc.ProcessPasswordResetEmail(t.Context(), user.Email))
			}
			require.Len(t, token, 64)
			require.Len(t, f.sender.messages, 1)
			message := f.sender.messages[0]
			require.Equal(t, []string{user.Email}, message.To)
			for _, line := range strings.Split(message.Text, "\n") {
				if !strings.HasPrefix(line, "https://") {
					continue
				}
				link, err := url.Parse(line)
				require.NoError(t, err)
				require.Empty(t, link.RawQuery)
				fragment, err := url.ParseQuery(link.Fragment)
				require.NoError(t, err)
				require.Equal(t, token, fragment.Get("token"))
				require.Equal(t, "app.example", link.Host)
				require.Contains(t, link.Path, "/base/")
			}
			require.Contains(t, message.HTML, token)
		})
	}
}

func TestAuthEmailProcessingFailures(t *testing.T) {
	t.Parallel()
	for _, purpose := range []string{"verification", "reset"} {
		for _, stage := range []string{"unknown", "inactive", "lookup", "delete", "create", "send", "missing repository", "invalid URL", "invalid TTL"} {
			t.Run(purpose+"/"+stage, func(t *testing.T) {
				t.Parallel()
				f := newAuthEmailFixture(t)
				user := &entity.User{ID: 7, Email: "user@example.com", Active: stage != "inactive"}
				failure := errors.New("controlled failure")
				var lookupErr error
				if stage == "unknown" {
					user, lookupErr = nil, repo.ErrNotFound
				}
				if stage == "lookup" {
					lookupErr = failure
				}
				if stage != "missing repository" {
					if purpose == "verification" {
						f.users.EXPECT().GetByID(gomock.Any(), uint(7)).Return(user, lookupErr)
					} else {
						f.users.EXPECT().GetByEmail(gomock.Any(), "user@example.com").Return(user, lookupErr)
					}
				}
				if stage == "missing repository" {
					f.uc.WithEmailVerification(nil, auth.VerificationConfig{Enabled: true}).WithPasswordReset(nil, auth.ResetConfig{})
				}
				if stage == "invalid URL" || stage == "invalid TTL" {
					base, ttl := "http://app.example", time.Hour
					if stage == "invalid TTL" {
						base, ttl = "https://app.example", 0
					}
					f.uc.WithEmailVerification(f.verify, auth.VerificationConfig{Enabled: true, TokenTTL: ttl, BaseURL: base}).
						WithPasswordReset(f.resets, auth.ResetConfig{TokenTTL: ttl, BaseURL: base})
				}
				if stage == "delete" || stage == "create" || stage == "send" {
					var deleteErr, createErr error
					if stage == "delete" {
						deleteErr = failure
					}
					if stage == "create" {
						createErr = failure
					}
					f.expectPersistence(purpose, deleteErr, createErr)
					if stage == "send" {
						f.sender.err = failure
					}
				}
				var err error
				if purpose == "verification" {
					err = f.uc.ProcessVerificationEmail(t.Context(), 7)
				} else {
					err = f.uc.ProcessPasswordResetEmail(t.Context(), "user@example.com")
				}
				switch stage {
				case "unknown", "inactive":
					require.NoError(t, err)
					require.Empty(t, f.sender.messages)
				case "missing repository", "invalid URL", "invalid TTL":
					require.ErrorIs(t, err, auth.ErrEmailUnavailable)
				default:
					require.ErrorIs(t, err, failure)
				}
			})
		}
	}
}

func TestAuthEmailRequestConfiguration(t *testing.T) {
	t.Parallel()
	f := newAuthEmailFixture(t)
	for _, address := range []string{"known@example.com", "unknown@example.com"} {
		require.ErrorIs(t, f.uc.RequestPasswordReset(t.Context(), address), auth.ErrEmailUnavailable)
	}
	f.uc.WithEmailVerification(f.verify, auth.VerificationConfig{Enabled: false})
	require.ErrorIs(t, f.uc.SendVerificationEmail(t.Context(), 7), auth.ErrVerificationDisabled)
	require.ErrorIs(t, f.uc.ResendVerification(t.Context(), "user@example.com"), auth.ErrVerificationDisabled)
	require.NoError(t, f.uc.ProcessVerificationEmail(t.Context(), 7))
	f.uc.WithEmailVerification(f.verify, auth.VerificationConfig{Enabled: true, AutoVerify: true})
	require.ErrorIs(t, f.uc.SendVerificationEmail(t.Context(), 7), auth.ErrEmailUnavailable)
	f.uc.WithEmailDelivery(false, false)
	f.users.EXPECT().GetByID(gomock.Any(), uint(7)).Return(&entity.User{ID: 7}, nil)
	f.users.EXPECT().SetEmailVerified(gomock.Any(), uint(7), gomock.Any()).Return(nil)
	require.NoError(t, f.uc.SendVerificationEmail(t.Context(), 7))
	require.Empty(t, f.sender.messages)
}
