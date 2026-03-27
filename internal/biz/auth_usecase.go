package biz

import (
	"context"
	"demo/internal/conf"
	"demo/internal/repo"
	"errors"
	"math/rand"
	"net"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthUseCase handles authentication business logic
type AuthUseCase struct {
	repo            repo.UserRepo
	loginLogRepo    repo.UserLoginLogRepo
	captchaRepo     repo.CaptchaRepo
	captchaConfig   *conf.Captcha
	log             *log.Helper
	jwtSecret       []byte
	jwtExpireDur    time.Duration
	parsedWhitelist []net.IPNet
}

// NewAuthUseCase creates a new AuthUseCase
func NewAuthUseCase(repo repo.UserRepo, loginLogRepo repo.UserLoginLogRepo, captchaRepo repo.CaptchaRepo,
	captchaConfig *conf.Captcha, logger log.Logger, jwtSecret string) *AuthUseCase {

	// Pre-parse IP whitelist CIDR at startup
	parsedWhitelist := make([]net.IPNet, 0, len(captchaConfig.IPWhitelist))
	logHelper := log.NewHelper(logger)
	for _, cidrStr := range captchaConfig.IPWhitelist {
		if _, ipNet, err := net.ParseCIDR(cidrStr); err == nil {
			parsedWhitelist = append(parsedWhitelist, *ipNet)
		} else if ip := net.ParseIP(cidrStr); ip != nil {
			// Single IP, convert to /32 (IPv4) or /128 (IPv6)
			var mask net.IPMask
			if ip.To4() != nil {
				mask = net.CIDRMask(32, 32)
			} else {
				mask = net.CIDRMask(128, 128)
			}
			parsedWhitelist = append(parsedWhitelist, net.IPNet{IP: ip, Mask: mask})
		} else {
			// Invalid IP/CIDR, log warning but continue startup
			logHelper.Warnf("invalid IP/CIDR in captcha ip_whitelist: %s, skipped", cidrStr)
		}
	}

	uc := &AuthUseCase{
		repo:            repo,
		loginLogRepo:    loginLogRepo,
		captchaRepo:     captchaRepo,
		captchaConfig:   captchaConfig,
		log:             log.NewHelper(logger),
		jwtSecret:       []byte(jwtSecret),
		jwtExpireDur:    24 * time.Hour,
		parsedWhitelist: parsedWhitelist,
	}

	// Set default values if not configured
	if uc.captchaConfig.Threshold <= 0 {
		uc.captchaConfig.Threshold = 3
	}
	if uc.captchaConfig.WindowMinutes <= 0 {
		uc.captchaConfig.WindowMinutes = 15
	}
	if uc.captchaConfig.ExpireMinutes <= 0 {
		uc.captchaConfig.ExpireMinutes = 5
	}

	return uc
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

// GenerateCaptcha generates a new captcha
// Returns captchaId, captchaCode, error
func (uc *AuthUseCase) GenerateCaptcha(ctx context.Context) (string, string, error) {
	if !uc.captchaConfig.Enabled {
		return "", "", errors.New("captcha is disabled")
	}

	captchaId := uuid.New().String()
	captchaCode := uc.generateCaptchaCode()

	err := uc.captchaRepo.StoreCaptcha(ctx, captchaId, captchaCode, uc.captchaConfig.ExpireMinutes)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("failed to store captcha: %v", err)
		return "", "", err
	}

	uc.log.WithContext(ctx).Debugf("generated new captcha: %s", captchaId)
	return captchaId, captchaCode, nil
}

// Login handles user login with captcha verification when needed
// Steps:
// 1. Check if captcha is required (based on failure count and IP whitelist)
// 2. If required, verify captcha before proceeding
// 3. Find user by username or email
// 4. Verify password
// 5. Generate JWT token
// 6. Record login log
// 7. Reset failure count on success, increment on failure
// 8. Return user info and token
func (uc *AuthUseCase) Login(ctx context.Context, loginBy string, isUsername bool, password string,
	clientIP, userAgent string, captchaID, captchaCode string) (*repo.User, string, time.Time, error) {

	var user *repo.User
	var err error
	var loginStatus int16
	var failReason string
	var loginUserID int64
	loginTime := time.Now()

	// Check if captcha is required
	requireCaptcha, err := uc.needCaptcha(ctx, clientIP, loginBy)
	if err != nil {
		uc.log.WithContext(ctx).Warnf("failed to check captcha requirement: %v", err)
		// Continue without captcha on Redis error, don't block login
		requireCaptcha = false
	}

	// Verify captcha if required
	if requireCaptcha {
		valid, verifyErr := uc.verifyCaptcha(ctx, captchaID, captchaCode)
		if verifyErr != nil {
			if errors.Is(verifyErr, errCaptchaNotFound) {
				return nil, "", time.Time{}, errors.New("captcha has expired, please refresh and try again")
			}
			if errors.Is(verifyErr, errCaptchaIncorrect) {
				return nil, "", time.Time{}, errors.New("captcha is incorrect, please try again")
			}
			return nil, "", time.Time{}, verifyErr
		}
		if !valid {
			// According to requirement: captcha incorrect doesn't increment failure count
			return nil, "", time.Time{}, errors.New("captcha is incorrect, please try again")
		}
		// Delete used captcha
		_ = uc.captchaRepo.DeleteCaptcha(ctx, captchaID)
	}

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
		// Increment failure count
		_ = uc.incrementFailureCount(ctx, clientIP, loginBy)
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
		// Increment failure count
		_ = uc.incrementFailureCount(ctx, clientIP, loginBy)
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
		// Increment failure count
		_ = uc.incrementFailureCount(ctx, clientIP, loginBy)
		return nil, "", time.Time{}, err
	}

	// Extract token ID (we use the token string as tokenID for lookup)
	tokenIDStr := token
	loginStatus = repo.LoginStatusSuccess
	_ = uc.recordLoginLog(ctx, loginUserID, clientIP, userAgent, loginStatus, "", loginTime, &tokenIDStr)

	// Reset failure count on successful login
	_ = uc.resetFailureCount(ctx, clientIP, loginBy)

	uc.log.WithContext(ctx).Infof("user logged in: %s (ID: %d)", user.Username, user.ID)
	return user, token, expiresAt, nil
}

// isIPWhitelisted checks if the given IP is in the whitelist
func (uc *AuthUseCase) isIPWhitelisted(ipStr string) bool {
	if ipStr == "" || len(uc.parsedWhitelist) == 0 {
		return false
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	for _, ipNet := range uc.parsedWhitelist {
		if ipNet.Contains(ip) {
			return true
		}
	}

	return false
}

// needCaptcha determines if captcha is required for this login attempt
func (uc *AuthUseCase) needCaptcha(ctx context.Context, clientIP string, loginBy string) (bool, error) {
	if !uc.captchaConfig.Enabled {
		return false, nil
	}

	// IP in whitelist doesn't need captcha
	if clientIP != "" && uc.isIPWhitelisted(clientIP) {
		return false, nil
	}

	threshold := uc.captchaConfig.Threshold
	var ipCount, userCount int
	var err error

	// Get IP failure count
	if clientIP != "" {
		ipCount, err = uc.captchaRepo.GetFailureCount(ctx, "ip:"+clientIP)
		if err != nil {
			uc.log.WithContext(ctx).Warnf("failed to get IP failure count: %v", err)
			// Continue with user count only
		}
		if ipCount >= threshold {
			return true, nil
		}
	}

	// Get user failure count
	userCount, err = uc.captchaRepo.GetFailureCount(ctx, "user:"+loginBy)
	if err != nil {
		uc.log.WithContext(ctx).Warnf("failed to get user failure count: %v", err)
		// If IP count already checked and below threshold, but user count query failed, don't require captcha
		return false, err
	}

	if userCount >= threshold {
		return true, nil
	}

	return false, nil
}

var (
	errCaptchaNotFound  = errors.New("captcha not found")
	errCaptchaIncorrect = errors.New("captcha incorrect")
	errCaptchaRequired  = errors.New("captcha is required")
)

// verifyCaptcha verifies the captcha code
// Returns (valid, error)
func (uc *AuthUseCase) verifyCaptcha(ctx context.Context, captchaID, code string) (bool, error) {
	if captchaID == "" || code == "" {
		return false, errCaptchaRequired
	}

	storedCode, err := uc.captchaRepo.GetCaptcha(ctx, captchaID)
	if err != nil {
		return false, err
	}
	if storedCode == "" {
		return false, errCaptchaNotFound
	}

	// Case-insensitive comparison
	if len(storedCode) != len(code) {
		return false, errCaptchaIncorrect
	}

	lowerStored := toLowerASCII(storedCode)
	lowerInput := toLowerASCII(code)

	if lowerStored == lowerInput {
		return true, nil
	}

	return false, errCaptchaIncorrect
}

// incrementFailureCount increments failure count for both IP and user
func (uc *AuthUseCase) incrementFailureCount(ctx context.Context, clientIP string, loginBy string) error {
	expireMinutes := uc.captchaConfig.WindowMinutes

	if clientIP != "" {
		_, err := uc.captchaRepo.IncrementFailureCount(ctx, "ip:"+clientIP, expireMinutes)
		if err != nil {
			uc.log.WithContext(ctx).Warnf("failed to increment IP failure count: %v", err)
			return err
		}
	}

	_, err := uc.captchaRepo.IncrementFailureCount(ctx, "user:"+loginBy, expireMinutes)
	if err != nil {
		uc.log.WithContext(ctx).Warnf("failed to increment user failure count: %v", err)
		return err
	}

	return nil
}

// resetFailureCount resets failure count for both IP and user
func (uc *AuthUseCase) resetFailureCount(ctx context.Context, clientIP string, loginBy string) error {
	if clientIP != "" {
		err := uc.captchaRepo.ResetFailureCount(ctx, "ip:"+clientIP)
		if err != nil {
			uc.log.WithContext(ctx).Warnf("failed to reset IP failure count: %v", err)
		}
	}

	err := uc.captchaRepo.ResetFailureCount(ctx, "user:"+loginBy)
	if err != nil {
		uc.log.WithContext(ctx).Warnf("failed to reset user failure count: %v", err)
		return err
	}

	return nil
}

// generateCaptchaCode generates a 4-character captcha that contains at least one digit and one letter
func (uc *AuthUseCase) generateCaptchaCode() string {
	const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Try up to 10 times to get a valid captcha that has at least one digit and one letter
	for i := 0; i < 10; i++ {
		code := make([]byte, 4)
		hasDigit := false
		hasLetter := false

		for j := 0; j < 4; j++ {
			idx := rng.Intn(len(chars))
			c := chars[idx]
			code[j] = c
			if c >= '0' && c <= '9' {
				hasDigit = true
			} else {
				hasLetter = true
			}
		}

		if hasDigit && hasLetter {
			return string(code)
		}
	}

	// If after 10 attempts still not valid, force it
	code := make([]byte, 4)
	code[0] = '0' + byte(rng.Intn(10)) // digit
	code[1] = 'A' + byte(rng.Intn(26)) // letter
	for j := 2; j < 4; j++ {
		code[j] = chars[rng.Intn(len(chars))]
	}
	return string(code)
}

// toLowerASCII converts ASCII characters to lowercase, leaves non-ASCII unchanged
func toLowerASCII(s string) string {
	res := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			res[i] = c + 32
		} else {
			res[i] = c
		}
	}
	return string(res)
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
