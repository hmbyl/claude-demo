package repo

import (
	"context"
	"time"
)

// 登录状态常量
const (
	LoginStatusFailed  int16 = 0 // 登录失败
	LoginStatusSuccess int16 = 1 // 登录成功
)

// UserLoginLog is the domain entity representing a user login record
type UserLoginLog struct {
	ID          int64
	UserID      int64
	LoginIP     string
	UserAgent   string
	LoginStatus int16
	FailReason  string
	LoginTime   time.Time
	GeoLocation *string
	TokenID     *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserLoginLogRepo defines the repository interface for user login log data access
// This follows the DDD pattern: interface defined in repo, implemented in data
type UserLoginLogRepo interface {
	// Create saves a new login record
	Create(ctx context.Context, log *UserLoginLog) (*UserLoginLog, error)
	// FindByUserID finds all login records by user ID (paginated)
	FindByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*UserLoginLog, int64, error)
	// FindByTokenID finds a login record by token ID (requires scanning all shards)
	FindByTokenID(ctx context.Context, tokenID string) (*UserLoginLog, error)
}
