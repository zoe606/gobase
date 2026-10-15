// Package tasks defines background task types and handlers.
package tasks

// Task type constants.
const (
	TypeEmailNotification = "email:notification"
	TypeEmailVerification = "auth:email_verification"
	TypePasswordReset     = "auth:password_reset"
	TypeUserCleanup       = "user:cleanup"
	TypeImageProcessing   = "media:image_processing"
)
