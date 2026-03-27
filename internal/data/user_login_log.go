package data

import (
	"context"
	"fmt"
	"demo/internal/repo"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// userLoginLogPO is the GORM persistent object for UserLoginLog
// Maps to the tbl_user_login_logs_xx table in database (dynamic sharding by user_id)
type userLoginLogPO struct {
	ID          int64          `gorm:"primaryKey"`
	UserID      int64          `gorm:"column:user_id"`
	LoginIP     string         `gorm:"column:login_ip;size:50"`
	UserAgent   string         `gorm:"column:user_agent;size:500"`
	LoginStatus int16          `gorm:"column:login_status"`
	FailReason  string         `gorm:"column:fail_reason;size:255"`
	LoginTime   time.Time      `gorm:"column:login_time"`
	GeoLocation *string        `gorm:"column:geo_location;size:100"`
	TokenID     *string        `gorm:"column:token_id;size:100"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;column:deleted_at"`
}

// TableName 根据 userID 动态分表
// 分表算法: userID % 10 -> 0-9，格式化为两位数 00-09
func (po userLoginLogPO) TableName() string {
	if po.UserID == 0 {
		// 当 userID 为 0 时返回默认表名（用于无 userID 的查询场景）
		return "tbl_user_login_logs_00"
	}
	idx := po.UserID % 10
	return fmt.Sprintf("tbl_user_login_logs_%02d", idx)
}

// userLoginLogRepo implements repo.UserLoginLogRepo with dynamic sharding
type userLoginLogRepo struct {
	data *Data
	log  *log.Helper
}

// NewUserLoginLogRepo creates a new userLoginLogRepo
func NewUserLoginLogRepo(data *Data, logger log.Logger) repo.UserLoginLogRepo {
	return &userLoginLogRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// Create implements repo.UserLoginLogRepo.Create
// Sharding: routes to the correct table based on userID
func (r *userLoginLogRepo) Create(ctx context.Context, log *repo.UserLoginLog) (*repo.UserLoginLog, error) {
	po := &userLoginLogPO{
		UserID:      log.UserID,
		LoginIP:     log.LoginIP,
		UserAgent:   log.UserAgent,
		LoginStatus: log.LoginStatus,
		FailReason:  log.FailReason,
		LoginTime:   log.LoginTime,
		GeoLocation: log.GeoLocation,
		TokenID:     log.TokenID,
	}

	err := r.data.db.WithContext(ctx).Create(po).Error
	if err != nil {
		return nil, err
	}

	log.ID = po.ID
	log.CreatedAt = po.CreatedAt
	log.UpdatedAt = po.UpdatedAt
	return log, nil
}

// FindByUserID implements repo.UserLoginLogRepo.FindByUserID
// Sharding: userID is known, routes directly to the correct shard
func (r *userLoginLogRepo) FindByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*repo.UserLoginLog, int64, error) {
	var poList []*userLoginLogPO
	var total int64

	// Create PO with userID to get correct table name
	po := userLoginLogPO{UserID: userID}

	query := r.data.db.WithContext(ctx).Table(po.TableName()).Where("user_id = ? AND deleted_at IS NULL", userID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated data
	offset := (page - 1) * pageSize
	if err := query.Order("login_time DESC").Offset(offset).Limit(pageSize).Find(&poList).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain entities
	var result []*repo.UserLoginLog
	for _, poItem := range poList {
		result = append(result, &repo.UserLoginLog{
			ID:          poItem.ID,
			UserID:      poItem.UserID,
			LoginIP:     poItem.LoginIP,
			UserAgent:   poItem.UserAgent,
			LoginStatus: poItem.LoginStatus,
			FailReason:  poItem.FailReason,
			LoginTime:   poItem.LoginTime,
			GeoLocation: poItem.GeoLocation,
			TokenID:     poItem.TokenID,
			CreatedAt:   poItem.CreatedAt,
			UpdatedAt:   poItem.UpdatedAt,
		})
	}

	return result, total, nil
}

// FindByTokenID implements repo.UserLoginLogRepo.FindByTokenID
// Sharding: tokenID doesn't contain userID, need to search all 10 shards sequentially
func (r *userLoginLogRepo) FindByTokenID(ctx context.Context, tokenID string) (*repo.UserLoginLog, error) {
	for i := 0; i < 10; i++ {
		tableName := fmt.Sprintf("tbl_user_login_logs_%02d", i)
		var po userLoginLogPO
		err := r.data.db.WithContext(ctx).Table(tableName).Where("token_id = ? AND deleted_at IS NULL", tokenID).First(&po).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				continue
			}
			return nil, err
		}
		// Found in current shard
		return &repo.UserLoginLog{
			ID:          po.ID,
			UserID:      po.UserID,
			LoginIP:     po.LoginIP,
			UserAgent:   po.UserAgent,
			LoginStatus: po.LoginStatus,
			FailReason:  po.FailReason,
			LoginTime:   po.LoginTime,
			GeoLocation: po.GeoLocation,
			TokenID:     po.TokenID,
			CreatedAt:   po.CreatedAt,
			UpdatedAt:   po.UpdatedAt,
		}, nil
	}

	// Not found in any shard
	return nil, nil
}
