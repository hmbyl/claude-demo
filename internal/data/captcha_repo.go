package data

import (
	"context"
	"demo/internal/repo"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
)

// captchaRepo implements repo.CaptchaRepo based on Redis
type captchaRepo struct {
	data *Data
	log  *log.Helper
}

// NewCaptchaRepo creates a new captchaRepo
func NewCaptchaRepo(data *Data, logger log.Logger) repo.CaptchaRepo {
	return &captchaRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// Redis key prefix constants
const (
	keyPrefixCaptcha = "captcha:code:"
	keyPrefixFailure = "captcha:fail:"
)

// StoreCaptcha stores a captcha in Redis with expiration
func (r *captchaRepo) StoreCaptcha(ctx context.Context, captchaId, code string, expireMinutes int) error {
	key := keyPrefixCaptcha + captchaId
	expire := time.Duration(expireMinutes) * time.Minute
	return r.data.rdb.Set(ctx, key, code, expire).Err()
}

// GetCaptcha retrieves a captcha from Redis
func (r *captchaRepo) GetCaptcha(ctx context.Context, captchaId string) (string, error) {
	key := keyPrefixCaptcha + captchaId
	code, err := r.data.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // not found
	}
	if err != nil {
		return "", err
	}
	return code, nil
}

// DeleteCaptcha deletes a captcha from Redis
func (r *captchaRepo) DeleteCaptcha(ctx context.Context, captchaId string) error {
	key := keyPrefixCaptcha + captchaId
	return r.data.rdb.Del(ctx, key).Err()
}

// GetFailureCount gets current failure count for the given key
func (r *captchaRepo) GetFailureCount(ctx context.Context, key string) (int, error) {
	fullKey := keyPrefixFailure + key
	count, err := r.data.rdb.Get(ctx, fullKey).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return count, nil
}

// IncrementFailureCount increments the failure count
// If the key doesn't exist, it's created with expiration
// Returns the new count after increment
func (r *captchaRepo) IncrementFailureCount(ctx context.Context, key string, expireMinutes int) (int, error) {
	fullKey := keyPrefixFailure + key
	count, err := r.data.rdb.Incr(ctx, fullKey).Result()
	if err != nil {
		return 0, err
	}

	// If this is the first increment, set expiration
	if count == 1 {
		expire := time.Duration(expireMinutes) * time.Minute
		_ = r.data.rdb.Expire(ctx, fullKey, expire)
	}

	return int(count), nil
}

// ResetFailureCount resets the failure count to zero (deletes the key)
func (r *captchaRepo) ResetFailureCount(ctx context.Context, key string) error {
	fullKey := keyPrefixFailure + key
	return r.data.rdb.Del(ctx, fullKey).Err()
}
