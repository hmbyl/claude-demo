package repo

import (
	"context"
	"time"
)

// User is the domain entity representing a user
type User struct {
	ID        int64
	Username  string
	Email     string
	Password  string // hashed password
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserRepo defines the repository interface for user data access
// This follows the DDD pattern: interface defined in repo, implemented in data
type UserRepo interface {
	// Create saves a new user
	Create(ctx context.Context, user *User) (*User, error)
	// FindByUsername finds a user by username
	FindByUsername(ctx context.Context, username string) (*User, error)
	// FindByEmail finds a user by email
	FindByEmail(ctx context.Context, email string) (*User, error)
}
