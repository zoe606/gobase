package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	libasynq "github.com/hibiken/asynq"

	internalemail "go-boilerplate/internal/email"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/repo/webapi/email"
	"go-boilerplate/internal/worker/tasks"
	"go-boilerplate/pkg/json"
	"go-boilerplate/pkg/tx"
)

func (uc *UseCase) WithEmailDelivery(configured, production bool) *UseCase {
	uc.emailDeliveryReady, uc.production = configured, production
	return uc
}

func (uc *UseCase) WithEmailSender(sender email.Sender) *UseCase {
	uc.mailer = sender
	return uc
}

func (uc *UseCase) WithTransactions(helper tx.Helper) *UseCase {
	uc.transactions = helper
	return uc
}

func (uc *UseCase) inTransaction(ctx context.Context, fn func(context.Context) error) error {
	if uc.transactions == nil {
		return fn(ctx)
	}
	return uc.transactions.RunInTx(ctx, fn)
}

func (uc *UseCase) enqueueAuthEmail(ctx context.Context, kind string, payload tasks.AuthEmailPayload) error {
	if !uc.emailDeliveryReady || uc.asynqClient == nil {
		return ErrEmailUnavailable
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode auth email: %w", err)
	}
	// Reuse the task ID within each calendar minute to reduce repeated token emails.
	digest := sha256.Sum256(data)
	id := kind + ":" + hex.EncodeToString(digest[:]) + ":" + strconv.FormatInt(time.Now().Unix()/60, 10)
	options := []libasynq.Option{libasynq.Queue("default"), libasynq.TaskID(id), libasynq.Retention(time.Minute)}
	options = append(options, uc.taskOptions...)
	_, err = uc.asynqClient.EnqueueContext(ctx, libasynq.NewTask(kind, data), options...)
	if errors.Is(err, libasynq.ErrTaskIDConflict) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("enqueue auth email: %w", errors.Join(ErrEmailUnavailable, err))
	}
	return nil
}

func (uc *UseCase) emailLink(base, path, token string) (string, error) {
	u, err := url.Parse(base)
	if err != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" ||
		(u.Scheme != "https" && (u.Scheme != "http" || uc.production)) {
		return "", ErrEmailUnavailable
	}
	u = u.JoinPath(path)
	u.Fragment = url.Values{"token": {token}}.Encode()
	return u.String(), nil
}

func generateSecureToken() (string, error) {
	var data [32]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(data[:]), nil
}

func (uc *UseCase) ProcessVerificationEmail(ctx context.Context, userID uint) error {
	if !uc.verificationConfig.Enabled {
		return nil
	}
	if !uc.emailDeliveryReady || uc.emailVerificationRepo == nil || uc.mailer == nil {
		return ErrEmailUnavailable
	}
	user, err := uc.userRepo.GetByID(ctx, userID)
	if errors.Is(err, repo.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("verification user: %w", err)
	}
	if user.IsEmailVerified() || !user.Active {
		return nil
	}
	token, err := generateSecureToken()
	if err != nil {
		return fmt.Errorf("verification token: %w", err)
	}
	link, err := uc.emailLink(uc.verificationConfig.BaseURL, "verify-email", token)
	if err != nil {
		return err
	}
	html, text, err := internalemail.RenderAuthEmail("Verify your email", uc.appName, link)
	if err != nil {
		return fmt.Errorf("verification template: %w", err)
	}
	if uc.verificationConfig.TokenTTL <= 0 {
		return ErrEmailUnavailable
	}
	if err := uc.inTransaction(ctx, func(ctx context.Context) error {
		if err := uc.emailVerificationRepo.DeleteByUserID(ctx, user.ID); err != nil {
			return err
		}
		return uc.emailVerificationRepo.Create(ctx, &entity.EmailVerification{
			UserID: user.ID, Token: token, ExpiresAt: time.Now().Add(uc.verificationConfig.TokenTTL),
		})
	}); err != nil {
		return fmt.Errorf("store verification token: %w", err)
	}
	return uc.mailer.Send(ctx, email.Email{To: []string{user.Email}, Subject: "Verify your email", HTML: html, Text: text})
}

func (uc *UseCase) ProcessPasswordResetEmail(ctx context.Context, address string) error {
	if !uc.emailDeliveryReady || uc.passwordResetRepo == nil || uc.mailer == nil {
		return ErrEmailUnavailable
	}
	user, err := uc.userRepo.GetByEmail(ctx, address)
	if errors.Is(err, repo.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("password reset user: %w", err)
	}
	if !user.Active {
		return nil
	}
	token, err := generateSecureToken()
	if err != nil {
		return fmt.Errorf("password reset token: %w", err)
	}
	link, err := uc.emailLink(uc.resetConfig.BaseURL, "reset-password", token)
	if err != nil {
		return err
	}
	html, text, err := internalemail.RenderAuthEmail("Reset your password", uc.appName, link)
	if err != nil {
		return fmt.Errorf("password reset template: %w", err)
	}
	if uc.resetConfig.TokenTTL <= 0 {
		return ErrEmailUnavailable
	}
	if err := uc.inTransaction(ctx, func(ctx context.Context) error {
		if err := uc.passwordResetRepo.DeleteByUserID(ctx, user.ID); err != nil {
			return err
		}
		return uc.passwordResetRepo.Create(ctx, &entity.PasswordReset{
			UserID: user.ID, Token: token, ExpiresAt: time.Now().Add(uc.resetConfig.TokenTTL),
		})
	}); err != nil {
		return fmt.Errorf("store password reset token: %w", err)
	}
	return uc.mailer.Send(ctx, email.Email{To: []string{user.Email}, Subject: "Reset your password", HTML: html, Text: text})
}
