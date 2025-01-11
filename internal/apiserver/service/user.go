package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/xmualex2023/i18n-translation/internal/apiserver/model"
	"github.com/xmualex2023/i18n-translation/internal/pkg/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrUserExists        = errors.New("user already exists")
	ErrInvalidCredential = errors.New("invalid username or password")
)

// Register
func (s *Service) Register(ctx context.Context, req *model.RegisterRequest) error {
	existingUser, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return fmt.Errorf("failed to get user, username: %s, error: %w", req.Username, err)
	}

	if existingUser != nil {
		return ErrUserExists
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return err
	}

	// create user
	user := &model.User{
		ID:       primitive.NewObjectID(),
		Username: req.Username,
		Password: hashedPassword,
	}

	return s.repo.CreateUser(ctx, user)
}

// Login
func (s *Service) Login(ctx context.Context, req *model.LoginRequest) (*model.TokenResponse, error) {
	user, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredential
	}

	if err := auth.CheckPassword(req.Password, user.Password); err != nil {
		return nil, ErrInvalidCredential
	}

	// create jwt token
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

// RefreshToken
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
