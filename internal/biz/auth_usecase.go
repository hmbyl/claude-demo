package biz

import (
	"context"
	"demo/internal/repo"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthUseCase handles authentication business logic
type AuthUseCase struct {
	repo         repo.UserRepo
	loginLogRepo repo.UserLoginLogRepo
	log          *log.Helper
	jwtSecret    []byte
	jwtExpireDur time.Duration
}

// NewAuthUseCase creates a new AuthUseCase
func NewAuthUseCase(repo repo.UserRepo, loginLogRepo repo.UserLoginLogRepo, logger log.Logger, jwtSecret string) *AuthUseCase {
	return &AuthUseCase{
		repo:         repo,
		loginLogRepo: loginLogRepo,
		log:          log.NewHelper(logger),
		jwtSecret:    []byte(jwtSecret),
		jwtExpireDur: 24 * time.Hour, // 24 hours expiration
	}
}

// Register handles user registration
// Steps:
// 1. Check if username already exists
// 2. Check if email already exists
// 3. Hash password with bcrypt
// 4. Save user to database
// 5. Generate JWT token
// 6. Return user info and token
func (uc *AuthUseCase) Register(ctx context.Context, username, email, password string) (*repo.User, string, time.Time, error) {
	// Check if username exists
	if existing, _ := uc.repo.FindByUsername(ctx, username); existing != nil {
		uc.log.WithContext(ctx).Infof("username %s already exists", username)
		return nil, "", time.Time{}, errors.New("username already exists")
	}

	// Check if email exists
	if existing, _ := uc.repo.FindByEmail(ctx, email); existing != nil {
		uc.log.WithContext(ctx).Infof("email %s already exists", email)
		return nil, "", time.Time{}, errors.New("email already exists")
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("failed to hash password: %v", err)
		return nil, "", time.Time{}, err
	}

	// Create user
	user := &repo.User{
		Username: username,
		Email:    email,
		Password: string(hashed),
	}

	user, err = uc.repo.Create(ctx, user)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("failed to create user: %v", err)
		return nil, "", time.Time{}, err
	}

	// Generate JWT
	expiresAt := time.Now().Add(uc.jwtExpireDur)
	token, err := uc.generateToken(user.ID, user.Username, expiresAt)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("failed to generate token: %v", err)
		return nil, "", time.Time{}, err
	}

	uc.log.WithContext(ctx).Infof("user registered: %s (ID: %d)", username, user.ID)
	return user, token, expiresAt, nil
}

// Login handles user login
// Steps:
// 1. Find user by username or email
// 2. Verify password
// 3. Generate JWT token
// 4. Record login log
// 5. Return user info and token
func (uc *AuthUseCase) Login(ctx context.Context, loginBy string, isUsername bool, password string, clientIP, userAgent string) (*repo.User, string, time.Time, error) {
	var user *repo.User
	var err error
	var loginStatus int16
	var failReason string
	var loginUserID int64
	loginTime := time.Now()

	// Find user
	if isUsername {
		user, err = uc.repo.FindByUsername(ctx, loginBy)
	} else {
		user, err = uc.repo.FindByEmail(ctx, loginBy)
	}

	if err != nil || user == nil {
		loginStatus = repo.LoginStatusFailed
		failReason = "user not found"
		// user unknown, userID = 0 -> shard 00
		_ = uc.recordLoginLog(ctx, 0, clientIP, userAgent, loginStatus, failReason, loginTime, nil)
		uc.log.WithContext(ctx).Infof("user not found: %s", loginBy)
		return nil, "", time.Time{}, errors.New("invalid credentials")
	}

	loginUserID = user.ID

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		loginStatus = repo.LoginStatusFailed
		failReason = "password mismatch"
		_ = uc.recordLoginLog(ctx, loginUserID, clientIP, userAgent, loginStatus, failReason, loginTime, nil)
		uc.log.WithContext(ctx).Infof("password mismatch for user: %s", loginBy)
		return nil, "", time.Time{}, errors.New("invalid credentials")
	}

	// Generate JWT
	expiresAt := time.Now().Add(uc.jwtExpireDur)
	token, err := uc.generateToken(user.ID, user.Username, expiresAt)
	if err != nil {
		loginStatus = repo.LoginStatusFailed
		failReason = "failed to generate token"
		_ = uc.recordLoginLog(ctx, loginUserID, clientIP, userAgent, loginStatus, failReason, loginTime, nil)
		uc.log.WithContext(ctx).Errorf("failed to generate token: %v", err)
		return nil, "", time.Time{}, err
	}

	// Extract token ID (we use the token string as tokenID for lookup)
	tokenIDStr := token
	loginStatus = repo.LoginStatusSuccess
	_ = uc.recordLoginLog(ctx, loginUserID, clientIP, userAgent, loginStatus, "", loginTime, &tokenIDStr)

	uc.log.WithContext(ctx).Infof("user logged in: %s (ID: %d)", user.Username, user.ID)
	return user, token, expiresAt, nil
}

// recordLoginLog records a login attempt to database
func (uc *AuthUseCase) recordLoginLog(ctx context.Context, userID int64, clientIP, userAgent string, status int16, failReason string, loginTime time.Time, tokenID *string) error {
	logEntry := &repo.UserLoginLog{
		UserID:      userID,
		LoginIP:     clientIP,
		UserAgent:   userAgent,
		LoginStatus: status,
		FailReason:  failReason,
		LoginTime:   loginTime,
		TokenID:     tokenID,
	}

	_, err := uc.loginLogRepo.Create(ctx, logEntry)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("failed to record login log: %v", err)
		return err
	}

	return nil
}

// generateToken generates a JWT token
func (uc *AuthUseCase) generateToken(userID int64, username string, expiresAt time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   userID,
		"username":  username,
		"exp":       expiresAt.Unix(),
		"issued_at": time.Now().Unix(),
	})

	return token.SignedString(uc.jwtSecret)
}
