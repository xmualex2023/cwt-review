package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrInvalidToken = errors.New("无效的令牌")
	ErrExpiredToken = errors.New("令牌已过期")
)

type Claims struct {
	UserID primitive.ObjectID `json:"user_id"`
	jwt.RegisteredClaims
}

// JWTMaker JWT 令牌生成器
type JWTMaker struct {
	secretKey []byte
	cache     TokenCache
}

func NewJWTMaker(secretKey string, cache TokenCache) *JWTMaker {
	return &JWTMaker{
		secretKey: []byte(secretKey),
		cache:     cache,
	}
}

// CreateToken 创建令牌
func (m *JWTMaker) CreateToken(ctx context.Context, userID primitive.ObjectID, duration time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().Add(duration)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", time.Time{}, err
	}

	// 缓存令牌
	if err := m.cache.Set(ctx, tokenString, claims); err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// VerifyToken 验证令牌
func (m *JWTMaker) VerifyToken(ctx context.Context, tokenString string) (*Claims, error) {
	// 首先从缓存中获取
	if claims, err := m.cache.Get(ctx, tokenString); err == nil {
		return claims, nil
	}

	// 缓存未命中，验证令牌
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// 缓存验证通过的令牌
	if err := m.cache.Set(ctx, tokenString, claims); err != nil {
		return nil, err
	}

	return claims, nil
}

// RevokeToken 吊销令牌
func (m *JWTMaker) RevokeToken(ctx context.Context, tokenString string) error {
	return m.cache.Delete(ctx, tokenString)
}

// RefreshToken 刷新令牌
func (m *JWTMaker) RefreshToken(ctx context.Context, oldToken string) (string, time.Time, error) {
	// 验证旧令牌
	claims, err := m.VerifyToken(ctx, oldToken)
	if err != nil {
		return "", time.Time{}, err
	}

	// 检查令牌是否即将过期（比如还有30%的有效期）
	now := time.Now()
	expiry := claims.ExpiresAt.Time
	threshold := expiry.Sub(claims.IssuedAt.Time) * 3 / 10

	if now.Add(threshold).Before(expiry) {
		return "", time.Time{}, errors.New("令牌尚未到达刷新时间")
	}

	// 吊销旧令牌
	if err := m.RevokeToken(ctx, oldToken); err != nil {
		return "", time.Time{}, err
	}

	// 创建新令牌
	return m.CreateToken(ctx, claims.UserID, m.cache.(*RedisTokenCache).defaultExpiry)
}
