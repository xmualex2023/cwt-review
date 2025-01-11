package service

import (
	"context"
	"errors"

	"github.com/xmualex2023/i18n-translation/internal/apiserver/model"
	"github.com/xmualex2023/i18n-translation/internal/pkg/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrUserExists        = errors.New("用户已存在")
	ErrInvalidCredential = errors.New("用户名或密码错误")
)

// Register 用户注册
func (s *Service) Register(ctx context.Context, req *model.RegisterRequest) error {
	// 检查用户是否已存在
	existingUser, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return ErrUserExists
	}

	// 对密码进行哈希处理
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return err
	}

	// 创建新用户
	user := &model.User{
		ID:       primitive.NewObjectID(),
		Username: req.Username,
		Password: hashedPassword,
	}

	return s.repo.CreateUser(ctx, user)
}

// Login 用户登录
func (s *Service) Login(ctx context.Context, req *model.LoginRequest) (*model.TokenResponse, error) {
	// 获取用户信息
	user, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredential
	}

	// 验证密码
	if err := auth.CheckPassword(req.Password, user.Password); err != nil {
		return nil, ErrInvalidCredential
	}

	// 创建 JWT 令牌
	jwtMaker := auth.NewJWTMaker(s.cfg.JWT.Secret, s.cache)
	token, expiresAt, err := jwtMaker.CreateToken(ctx, user.ID, s.cfg.JWT.Expire)
	if err != nil {
		return nil, err
	}

	return &model.TokenResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// RefreshToken 刷新访问令牌
func (s *Service) RefreshToken(ctx context.Context, oldToken string) (*model.TokenResponse, error) {
	jwtMaker := auth.NewJWTMaker(s.cfg.JWT.Secret, s.cache)
	token, expiresAt, err := jwtMaker.RefreshToken(ctx, oldToken)
	if err != nil {
		return nil, err
	}

	return &model.TokenResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}
