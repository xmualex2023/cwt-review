package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TODO: 可以在当前基础上，再增加一层内存缓存（local cache）
// TokenCache 令牌缓存接口
type TokenCache interface {
	// Set 缓存令牌
	Set(ctx context.Context, token string, claims *Claims) error
	// Get 获取令牌信息
	Get(ctx context.Context, token string) (*Claims, error)
	// Delete 删除令牌
	Delete(ctx context.Context, token string) error
}

// RedisTokenCache Redis 令牌缓存实现
type RedisTokenCache struct {
	client        *redis.Client
	keyPrefix     string
	defaultExpiry time.Duration
}

func NewRedisTokenCache(client *redis.Client, keyPrefix string, defaultExpiry time.Duration) *RedisTokenCache {
	return &RedisTokenCache{
		client:        client,
		keyPrefix:     keyPrefix,
		defaultExpiry: defaultExpiry,
	}
}

func (c *RedisTokenCache) tokenKey(token string) string {
	return fmt.Sprintf("%s:%s", c.keyPrefix, token)
}

func (c *RedisTokenCache) Set(ctx context.Context, token string, claims *Claims) error {
	data, err := json.Marshal(claims)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, c.tokenKey(token), data, c.defaultExpiry).Err()
}

func (c *RedisTokenCache) Get(ctx context.Context, token string) (*Claims, error) {
	data, err := c.client.Get(ctx, c.tokenKey(token)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	var claims Claims
	if err := json.Unmarshal(data, &claims); err != nil {
		return nil, err
	}

	return &claims, nil
}

func (c *RedisTokenCache) Delete(ctx context.Context, token string) error {
	return c.client.Del(ctx, c.tokenKey(token)).Err()
}
