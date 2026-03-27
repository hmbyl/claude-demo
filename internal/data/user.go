package data

import (
	"context"
	"demo/internal/repo"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// userPO is the GORM persistent object for User
// Maps to the users table in database
type userPO struct {
	ID        int64 `gorm:"primaryKey"`
	Username  string `gorm:"uniqueIndex;size:50"`
	Email     string `gorm:"uniqueIndex;size:100"`
	Password  string `gorm:"column:password_hash;size:255"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName specifies the table name
func (userPO) TableName() string {
	return "users"
}

// userRepo implements repo.UserRepo
type userRepo struct {
	data *Data
	log  *log.Helper
}

// NewUserRepo creates a new userRepo
func NewUserRepo(data *Data, logger log.Logger) repo.UserRepo {
	return &userRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// Create implements repo.UserRepo.Create
func (r *userRepo) Create(ctx context.Context, user *repo.User) (*repo.User, error) {
	po := &userPO{
		Username:  user.Username,
		Email:     user.Email,
		Password:  user.Password,
	}

	err := r.data.db.WithContext(ctx).Create(po).Error
	if err != nil {
		return nil, err
	}

	user.ID = po.ID
	user.CreatedAt = po.CreatedAt
	user.UpdatedAt = po.UpdatedAt
	return user, nil
}

// FindByUsername implements repo.UserRepo.FindByUsername
func (r *userRepo) FindByUsername(ctx context.Context, username string) (*repo.User, error) {
	var po userPO
	err := r.data.db.WithContext(ctx).Where("username = ?", username).First(&po).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &repo.User{
		ID:        po.ID,
		Username:  po.Username,
		Email:     po.Email,
		Password:  po.Password,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}, nil
}

// FindByEmail implements repo.UserRepo.FindByEmail
func (r *userRepo) FindByEmail(ctx context.Context, email string) (*repo.User, error) {
	var po userPO
	err := r.data.db.WithContext(ctx).Where("email = ?", email).First(&po).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &repo.User{
		ID:        po.ID,
		Username:  po.Username,
		Email:     po.Email,
		Password:  po.Password,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}, nil
}
