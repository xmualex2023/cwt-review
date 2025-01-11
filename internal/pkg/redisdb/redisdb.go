package redisdb

import "github.com/redis/go-redis/v9"

// 这里封装了redis client 接口，方便其他模块使用

type IRedisClient interface {
}

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(opt *redis.Options) *RedisClient {
	return &RedisClient{
		client: redis.NewClient(opt),
	}
}
