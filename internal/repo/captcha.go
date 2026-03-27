package repo

import (
	"context"
)

// CaptchaRepo defines the interface for captcha storage and failure count operations
// This follows DDD: interface defined in domain (repo), implemented in data layer
type CaptchaRepo interface {
	// StoreCaptcha stores a captcha in Redis with expiration
	// captchaId: unique identifier for the captcha
	// code: captcha content
	// expireMinutes: expiration time in minutes
	StoreCaptcha(ctx context.Context, captchaId, code string, expireMinutes int) error

	// GetCaptcha retrieves a captcha from Redis
	// Returns the stored code or error if not found
	GetCaptcha(ctx context.Context, captchaId string) (string, error)

	// DeleteCaptcha deletes a captcha from Redis after verification
	DeleteCaptcha(ctx context.Context, captchaId string) error

	// GetFailureCount gets current failure count for the given key
	// Key can be IP or username/email
	GetFailureCount(ctx context.Context, key string) (int, error)

	// IncrementFailureCount increments the failure count
	// If the key doesn't exist, it's created with expiration
	// Returns the new count after increment
	IncrementFailureCount(ctx context.Context, key string, expireMinutes int) (int, error)

	// ResetFailureCount resets the failure count to zero (deletes the key)
	ResetFailureCount(ctx context.Context, key string) error
}
