package service

import (
	"context"
	pb "demo/api/auth/v1"
	"demo/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// AuthService implements the auth v1 Auth service
type AuthService struct {
	pb.UnimplementedAuthServer
	uc  *biz.AuthUseCase
	log *log.Helper
}

// NewAuthService creates a new AuthService
func NewAuthService(uc *biz.AuthUseCase, logger log.Logger) *AuthService {
	return &AuthService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// Register handles user registration
func (s *AuthService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterReply, error) {
	user, token, expiresAt, err := s.uc.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &pb.RegisterReply{
		UserId:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		AccessToken: token,
		ExpiresAt:   expiresAt.Unix(),
	}, nil
}

// Login handles user login
func (s *AuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	var (
		loginBy    string
		isUsername bool
	)

	if req.GetUsername() != "" {
		loginBy = req.GetUsername()
		isUsername = true
	} else {
		loginBy = req.GetEmail()
		isUsername = false
	}

	// For now, clientIP and userAgent are left empty
	// In production, these should be extracted from HTTP request context by the server
	captchaID := req.GetCaptchaId()
	captchaCode := req.GetCaptchaCode()
	user, token, expiresAt, err := s.uc.Login(ctx, loginBy, isUsername, req.Password, "", "", captchaID, captchaCode)
	if err != nil {
		return nil, err
	}

	return &pb.LoginReply{
		UserId:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		AccessToken: token,
		ExpiresAt:   expiresAt.Unix(),
	}, nil
}

// GetCaptcha gets a new captcha
func (s *AuthService) GetCaptcha(ctx context.Context, req *pb.GetCaptchaRequest) (*pb.GetCaptchaReply, error) {
	captchaID, captchaCode, expireAt, err := s.uc.GenerateCaptcha(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.GetCaptchaReply{
		CaptchaId:   captchaID,
		CaptchaCode: captchaCode,
		ExpireAt:    expireAt.Unix(),
	}, nil
}
