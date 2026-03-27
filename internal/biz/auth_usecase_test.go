package biz

import (
	"context"
	"demo/internal/conf"
	"demo/internal/repo"
	"errors"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepo mocks repo.UserRepo for testing
type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) FindByUsername(ctx context.Context, username string) (*repo.User, error) {
	args := m.Called(ctx, username)
	if user, ok := args.Get(0).(*repo.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (*repo.User, error) {
	args := m.Called(ctx, email)
	if user, ok := args.Get(0).(*repo.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepo) Create(ctx context.Context, user *repo.User) (*repo.User, error) {
	args := m.Called(ctx, user)
	if retUser, ok := args.Get(0).(*repo.User); ok {
		return retUser, args.Error(1)
	}
	return nil, args.Error(1)
}

// MockUserLoginLogRepo mocks repo.UserLoginLogRepo for testing
type MockUserLoginLogRepo struct {
	mock.Mock
}

func (m *MockUserLoginLogRepo) Create(ctx context.Context, log *repo.UserLoginLog) (*repo.UserLoginLog, error) {
	args := m.Called(ctx, log)
	if retLog, ok := args.Get(0).(*repo.UserLoginLog); ok {
		return retLog, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserLoginLogRepo) FindByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*repo.UserLoginLog, int64, error) {
	args := m.Called(ctx, userID, page, pageSize)
	if logs, ok := args.Get(0).([]*repo.UserLoginLog); ok {
		return logs, args.Get(1).(int64), args.Error(2)
	}
	return nil, 0, args.Error(2)
}

func (m *MockUserLoginLogRepo) FindByTokenID(ctx context.Context, tokenID string) (*repo.UserLoginLog, error) {
	args := m.Called(ctx, tokenID)
	if log, ok := args.Get(0).(*repo.UserLoginLog); ok {
		return log, args.Error(1)
	}
	return nil, args.Error(1)
}

// MockCaptchaRepo mocks repo.CaptchaRepo for testing
type MockCaptchaRepo struct {
	mock.Mock
}

func (m *MockCaptchaRepo) StoreCaptcha(ctx context.Context, captchaId, code string, expireMinutes int) error {
	args := m.Called(ctx, captchaId, code, expireMinutes)
	return args.Error(0)
}

func (m *MockCaptchaRepo) GetCaptcha(ctx context.Context, captchaId string) (string, error) {
	args := m.Called(ctx, captchaId)
	return args.String(0), args.Error(1)
}

func (m *MockCaptchaRepo) DeleteCaptcha(ctx context.Context, captchaId string) error {
	args := m.Called(ctx, captchaId)
	return args.Error(0)
}

func (m *MockCaptchaRepo) GetFailureCount(ctx context.Context, key string) (int, error) {
	args := m.Called(ctx, key)
	return args.Int(0), args.Error(1)
}

func (m *MockCaptchaRepo) IncrementFailureCount(ctx context.Context, key string, expireMinutes int) (int, error) {
	args := m.Called(ctx, key, expireMinutes)
	return args.Int(0), args.Error(1)
}

func (m *MockCaptchaRepo) ResetFailureCount(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func createTestUseCase(t *testing.T, mockUserRepo *MockUserRepo, mockLoginLogRepo *MockUserLoginLogRepo, mockCaptchaRepo *MockCaptchaRepo, captchaConfig *conf.Captcha) *AuthUseCase {
	t.Helper()
	logger := testLogger{}
	jwtSecret := "test-jwt-secret"
	return NewAuthUseCase(mockUserRepo, mockLoginLogRepo, mockCaptchaRepo, captchaConfig, logger, jwtSecret)
}

// testLogger is a no-op logger for testing
type testLogger struct{}

func (l testLogger) Log(level log.Level, keyvals ...interface{}) error {
	return nil
}

// TestLogin is a table-driven test for AuthUseCase.Login
func TestLogin(t *testing.T) {
	testUser := &repo.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password is "password"
	}

	defaultCaptchaConfig := &conf.Captcha{
		Enabled:       true,
		Threshold:     3,
		WindowMinutes: 15,
		ExpireMinutes: 5,
		IPWhitelist:   []string{},
	}

	tests := []struct {
		name          string
		loginBy       string
		isUsername    bool
		password      string
		clientIP      string
		captchaID     string
		captchaCode   string
		captchaConfig *conf.Captcha
		setupMocks    func(*MockUserRepo, *MockUserLoginLogRepo, *MockCaptchaRepo)
		wantUser      *repo.User
		wantErr       bool
		errContains   string
	}{
		{
			name:          "1. successful login when captcha not required",
			loginBy:       "testuser",
			isUsername:    true,
			password:      "password",
			clientIP:      "192.168.1.1",
			captchaID:     "",
			captchaCode:   "",
			captchaConfig: defaultCaptchaConfig,
			setupMocks: func(mur *MockUserRepo, mlr *MockUserLoginLogRepo, mcr *MockCaptchaRepo) {
				mur.On("FindByUsername", mock.Anything, "testuser").Return(testUser, nil)
				mcr.On("GetFailureCount", mock.Anything, "ip:192.168.1.1").Return(0, nil)
				mcr.On("GetFailureCount", mock.Anything, "user:testuser").Return(0, nil)
				mlr.On("Create", mock.Anything, mock.Anything).Return(&repo.UserLoginLog{}, nil)
				mcr.On("ResetFailureCount", mock.Anything, "ip:192.168.1.1").Return(nil)
				mcr.On("ResetFailureCount", mock.Anything, "user:testuser").Return(nil)
			},
			wantUser:    testUser,
			wantErr:     false,
			errContains: "",
		},
		{
			name:          "2. login fails when user not found and captcha not required",
			loginBy:       "nonexistent",
			isUsername:    true,
			password:      "password",
			clientIP:      "192.168.1.1",
			captchaID:     "",
			captchaCode:   "",
			captchaConfig: defaultCaptchaConfig,
			setupMocks: func(mur *MockUserRepo, mlr *MockUserLoginLogRepo, mcr *MockCaptchaRepo) {
				mur.On("FindByUsername", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))
				mcr.On("GetFailureCount", mock.Anything, "ip:192.168.1.1").Return(0, nil)
				mcr.On("GetFailureCount", mock.Anything, "user:nonexistent").Return(0, nil)
				mlr.On("Create", mock.Anything, mock.Anything).Return(&repo.UserLoginLog{}, nil)
				mcr.On("IncrementFailureCount", mock.Anything, "ip:192.168.1.1", 15).Return(1, nil)
				mcr.On("IncrementFailureCount", mock.Anything, "user:nonexistent", 15).Return(1, nil)
			},
			wantUser:    nil,
			wantErr:     true,
			errContains: "invalid credentials",
		},
		{
			name:          "3. login fails when password wrong and captcha not required",
			loginBy:       "testuser",
			isUsername:    true,
			password:      "wrongpassword",
			clientIP:      "192.168.1.1",
			captchaID:     "",
			captchaCode:   "",
			captchaConfig: defaultCaptchaConfig,
			setupMocks: func(mur *MockUserRepo, mlr *MockUserLoginLogRepo, mcr *MockCaptchaRepo) {
				mur.On("FindByUsername", mock.Anything, "testuser").Return(testUser, nil)
				mcr.On("GetFailureCount", mock.Anything, "ip:192.168.1.1").Return(0, nil)
				mcr.On("GetFailureCount", mock.Anything, "user:testuser").Return(0, nil)
				mlr.On("Create", mock.Anything, mock.Anything).Return(&repo.UserLoginLog{}, nil)
				mcr.On("IncrementFailureCount", mock.Anything, "ip:192.168.1.1", 15).Return(1, nil)
				mcr.On("IncrementFailureCount", mock.Anything, "user:testuser", 15).Return(1, nil)
			},
			wantUser:    nil,
			wantErr:     true,
			errContains: "invalid credentials",
		},
		{
			name:          "4. require captcha when failure count reaches threshold, but no captcha provided - should fail",
			loginBy:       "testuser",
			isUsername:    true,
			password:      "password",
			clientIP:      "192.168.1.1",
			captchaID:     "",
			captchaCode:   "",
			captchaConfig: defaultCaptchaConfig,
			setupMocks: func(mur *MockUserRepo, mlr *MockUserLoginLogRepo, mcr *MockCaptchaRepo) {
				mcr.On("GetFailureCount", mock.Anything, "ip:192.168.1.1").Return(3, nil)
				mcr.On("GetFailureCount", mock.Anything, "user:testuser").Return(2, nil)
				// No user lookup because we fail before that
			},
			wantUser:    nil,
			wantErr:     true,
			errContains: "captcha is required",
		},
		{
			name:          "5. require captcha, wrong captcha code - should fail, failure count not incremented",
			loginBy:       "testuser",
			isUsername:    true,
			password:      "password",
			clientIP:      "192.168.1.1",
			captchaID:     "test-captcha-id",
			captchaCode:   "wrong",
			captchaConfig: defaultCaptchaConfig,
			setupMocks: func(mur *MockUserRepo, mlr *MockUserLoginLogRepo, mcr *MockCaptchaRepo) {
				mcr.On("GetFailureCount", mock.Anything, "ip:192.168.1.1").Return(3, nil)
				mcr.On("GetFailureCount", mock.Anything, "user:testuser").Return(2, nil)
				mcr.On("GetCaptcha", mock.Anything, "test-captcha-id").Return("abcd", nil)
				// Should NOT increment failure count because captcha is wrong but password not checked yet
			},
			wantUser:    nil,
			wantErr:     true,
			errContains: "captcha is incorrect",
		},
		{
			name:          "6. require captcha, expired captcha - should fail",
			loginBy:       "testuser",
			isUsername:    true,
			password:      "password",
			clientIP:      "192.168.1.1",
			captchaID:     "expired-id",
			captchaCode:   "abcd",
			captchaConfig: defaultCaptchaConfig,
			setupMocks: func(mur *MockUserRepo, mlr *MockUserLoginLogRepo, mcr *MockCaptchaRepo) {
				mcr.On("GetFailureCount", mock.Anything, "ip:192.168.1.1").Return(3, nil)
				mcr.On("GetFailureCount", mock.Anything, "user:testuser").Return(2, nil)
				mcr.On("GetCaptcha", mock.Anything, "expired-id").Return("", nil)
			},
			wantUser:    nil,
			wantErr:     true,
			errContains: "captcha has expired",
		},
		{
			name:          "7. require captcha, correct captcha, successful login",
			loginBy:       "testuser",
			isUsername:    true,
			password:      "password",
			clientIP:      "192.168.1.1",
			captchaID:     "valid-id",
			captchaCode:   "abcd",
			captchaConfig: defaultCaptchaConfig,
			setupMocks: func(mur *MockUserRepo, mlr *MockUserLoginLogRepo, mcr *MockCaptchaRepo) {
				mcr.On("GetFailureCount", mock.Anything, "ip:192.168.1.1").Return(3, nil)
				mcr.On("GetFailureCount", mock.Anything, "user:testuser").Return(2, nil)
				mcr.On("GetCaptcha", mock.Anything, "valid-id").Return("abcd", nil)
				mcr.On("DeleteCaptcha", mock.Anything, "valid-id").Return(nil)
				mur.On("FindByUsername", mock.Anything, "testuser").Return(testUser, nil)
				mlr.On("Create", mock.Anything, mock.Anything).Return(&repo.UserLoginLog{}, nil)
				mcr.On("ResetFailureCount", mock.Anything, "ip:192.168.1.1").Return(nil)
				mcr.On("ResetFailureCount", mock.Anything, "user:testuser").Return(nil)
			},
			wantUser:    testUser,
			wantErr:     false,
			errContains: "",
		},
		{
			name:          "8. cannot get client IP (empty) - should use user dimension only",
			loginBy:       "testuser",
			isUsername:    true,
			password:      "password",
			clientIP:      "",
			captchaID:     "",
			captchaCode:   "",
			captchaConfig: defaultCaptchaConfig,
			setupMocks: func(mur *MockUserRepo, mlr *MockUserLoginLogRepo, mcr *MockCaptchaRepo) {
				// No IP, so only check user count
				mcr.On("GetFailureCount", mock.Anything, "user:testuser").Return(0, nil)
				mur.On("FindByUsername", mock.Anything, "testuser").Return(testUser, nil)
				mlr.On("Create", mock.Anything, mock.Anything).Return(&repo.UserLoginLog{}, nil)
				mcr.On("ResetFailureCount", mock.Anything, "user:testuser").Return(nil)
			},
			wantUser:    testUser,
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "9. IP is in whitelist - should skip captcha even when failure count is high",
			loginBy:     "testuser",
			isUsername:  true,
			password:    "password",
			clientIP:    "127.0.0.1",
			captchaID:   "",
			captchaCode: "",
			captchaConfig: &conf.Captcha{
				Enabled:       true,
				Threshold:     3,
				WindowMinutes: 15,
				ExpireMinutes: 5,
				IPWhitelist:   []string{"127.0.0.1/32", "192.168.0.0/16"},
			},
			setupMocks: func(mur *MockUserRepo, mlr *MockUserLoginLogRepo, mcr *MockCaptchaRepo) {
				// IP is whitelisted, should not check failure count
				mur.On("FindByUsername", mock.Anything, "testuser").Return(testUser, nil)
				mlr.On("Create", mock.Anything, mock.Anything).Return(&repo.UserLoginLog{}, nil)
				mcr.On("ResetFailureCount", mock.Anything, "ip:127.0.0.1").Return(nil)
				mcr.On("ResetFailureCount", mock.Anything, "user:testuser").Return(nil)
			},
			wantUser:    testUser,
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "10. captcha disabled - always skip captcha check",
			loginBy:     "testuser",
			isUsername:  true,
			password:    "password",
			clientIP:    "192.168.1.1",
			captchaID:   "",
			captchaCode: "",
			captchaConfig: &conf.Captcha{
				Enabled:       false,
				Threshold:     3,
				WindowMinutes: 15,
				ExpireMinutes: 5,
			},
			setupMocks: func(mur *MockUserRepo, mlr *MockUserLoginLogRepo, mcr *MockCaptchaRepo) {
				mur.On("FindByUsername", mock.Anything, "testuser").Return(testUser, nil)
				mlr.On("Create", mock.Anything, mock.Anything).Return(&repo.UserLoginLog{}, nil)
				mcr.On("ResetFailureCount", mock.Anything, "ip:192.168.1.1").Return(nil)
				mcr.On("ResetFailureCount", mock.Anything, "user:testuser").Return(nil)
			},
			wantUser:    testUser,
			wantErr:     false,
			errContains: "",
		},
		{
			name:          "11. DAO layer query timeout when getting user - should return error and increment failure",
			loginBy:       "testuser",
			isUsername:    true,
			password:      "password",
			clientIP:      "192.168.1.1",
			captchaID:     "",
			captchaCode:   "",
			captchaConfig: defaultCaptchaConfig,
			setupMocks: func(mur *MockUserRepo, mlr *MockUserLoginLogRepo, mcr *MockCaptchaRepo) {
				mcr.On("GetFailureCount", mock.Anything, "ip:192.168.1.1").Return(0, nil)
				mcr.On("GetFailureCount", mock.Anything, "user:testuser").Return(0, nil)
				mur.On("FindByUsername", mock.Anything, "testuser").Return(nil, context.DeadlineExceeded)
				mlr.On("Create", mock.Anything, mock.Anything).Return(&repo.UserLoginLog{}, nil)
				mcr.On("IncrementFailureCount", mock.Anything, "ip:192.168.1.1", 15).Return(1, nil)
				mcr.On("IncrementFailureCount", mock.Anything, "user:testuser", 15).Return(1, nil)
			},
			wantUser:    nil,
			wantErr:     true,
			errContains: "invalid credentials",
		},
		{
			name:          "12. loginBy email when IP failure count reaches threshold",
			loginBy:       "test@example.com",
			isUsername:    false,
			password:      "password",
			clientIP:      "10.0.0.1",
			captchaID:     "captcha-id",
			captchaCode:   "abcd",
			captchaConfig: defaultCaptchaConfig,
			setupMocks: func(mur *MockUserRepo, mlr *MockUserLoginLogRepo, mcr *MockCaptchaRepo) {
				mcr.On("GetFailureCount", mock.Anything, "ip:10.0.0.1").Return(3, nil)
				mcr.On("GetFailureCount", mock.Anything, "user:test@example.com").Return(1, nil)
				mcr.On("GetCaptcha", mock.Anything, "captcha-id").Return("abcd", nil)
				mcr.On("DeleteCaptcha", mock.Anything, "captcha-id").Return(nil)
				mur.On("FindByEmail", mock.Anything, "test@example.com").Return(testUser, nil)
				mlr.On("Create", mock.Anything, mock.Anything).Return(&repo.UserLoginLog{}, nil)
				mcr.On("ResetFailureCount", mock.Anything, "ip:10.0.0.1").Return(nil)
				mcr.On("ResetFailureCount", mock.Anything, "user:test@example.com").Return(nil)
			},
			wantUser:    testUser,
			wantErr:     false,
			errContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mur := new(MockUserRepo)
			mlr := new(MockUserLoginLogRepo)
			mcr := new(MockCaptchaRepo)
			if tt.setupMocks != nil {
				tt.setupMocks(mur, mlr, mcr)
			}

			uc := createTestUseCase(t, mur, mlr, mcr, tt.captchaConfig)
			ctx := context.Background()

			user, token, expiresAt, err := uc.Login(ctx, tt.loginBy, tt.isUsername, tt.password, tt.clientIP, "", tt.captchaID, tt.captchaCode)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				// Verify that mocks were called
				mur.AssertExpectations(t)
				mlr.AssertExpectations(t)
				mcr.AssertExpectations(t)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, user)
			assert.Equal(t, tt.wantUser.ID, user.ID)
			assert.Equal(t, tt.wantUser.Username, user.Username)
			assert.NotEmpty(t, token)
			assert.False(t, expiresAt.IsZero())

			// Verify JWT token is valid
			parser := jwt.NewParser()
			claims, err := parser.Parse(token, func(token *jwt.Token) (interface{}, error) {
				return []byte("test-jwt-secret"), nil
			})
			assert.NoError(t, err)
			assert.True(t, claims.Valid)

			mur.AssertExpectations(t)
			mlr.AssertExpectations(t)
			mcr.AssertExpectations(t)
		})
	}
}

// TestGenerateCaptchaCode tests that generated captcha always contains at least one digit and one letter
func TestGenerateCaptchaCode(t *testing.T) {
	uc := &AuthUseCase{}
	// Test many times to ensure the invariant holds
	for i := 0; i < 100; i++ {
		code := uc.generateCaptchaCode()
		assert.Len(t, code, 4)
		hasDigit := false
		hasLetter := false
		for _, c := range code {
			if c >= '0' && c <= '9' {
				hasDigit = true
			} else if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
				hasLetter = true
			}
		}
		assert.True(t, hasDigit, "captcha %q should contain at least one digit", code)
		assert.True(t, hasLetter, "captcha %q should contain at least one letter", code)
	}
}

// TestToLowerASCII tests case-insensitive comparison
func TestToLowerASCII(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ABCD", "abcd"},
		{"AbCd", "abcd"},
		{"1234", "1234"},
		{"aBc1", "abc1"},
	}
	for _, tt := range tests {
		result := toLowerASCII(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

// TestIPWhitelistMatching tests IP whitelist matching with CIDR
func TestIPWhitelistMatching(t *testing.T) {
	uc := NewAuthUseCase(nil, nil, nil, &conf.Captcha{
		IPWhitelist: []string{"127.0.0.1", "192.168.0.0/16", "10.0.0.0/8"},
	}, testLogger{}, "test")

	tests := []struct {
		ip   string
		want bool
	}{
		{"127.0.0.1", true},
		{"127.0.0.2", false},
		{"192.168.1.1", true},
		{"192.168.100.100", true},
		{"192.169.1.1", false},
		{"10.1.2.3", true},
		{"11.0.0.1", false},
		{"", false},
		{"invalid-ip", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			result := uc.isIPWhitelisted(tt.ip)
			assert.Equal(t, tt.want, result)
		})
	}
}
