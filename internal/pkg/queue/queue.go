package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xmualex2023/i18n-translation/internal/pkg/metrics"
)

// Task 任务接口
type Task interface {
	// GetID 获取任务ID
	GetID() string
}

// Queue 队列接口
type Queue interface {
	// Enqueue 将任务加入队列
	Enqueue(ctx context.Context, task Task) error
	// Dequeue 从队列中获取任务
	Dequeue(ctx context.Context) (Task, error)
}

// RedisQueue Redis队列实现
type RedisQueue struct {
	client     *redis.Client
	queueKey   string
	retryCount int
	retryDelay time.Duration
}

func NewRedisQueue(client *redis.Client, queueKey string) *RedisQueue {
	return &RedisQueue{
		client:     client,
		queueKey:   queueKey,
		retryCount: 3,
		retryDelay: 5 * time.Second,
	}
}

// Enqueue 将任务加入队列
func (q *RedisQueue) Enqueue(ctx context.Context, task Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	if err := q.client.LPush(ctx, q.queueKey, data).Err(); err != nil {
		return err
	}

	// 更新队列大小
	size, err := q.client.LLen(ctx, q.queueKey).Result()
	if err == nil {
		metrics.SetQueueSize(int(size))
	}

	return nil
}

// Dequeue 从队列中获取任务
func (q *RedisQueue) Dequeue(ctx context.Context) (Task, error) {
	result, err := q.client.BRPop(ctx, 0, q.queueKey).Result()
	if err != nil {
		return nil, err
	}

	// 更新队列大小
	size, err := q.client.LLen(ctx, q.queueKey).Result()
	if err == nil {
		metrics.SetQueueSize(int(size))
	}

	var task Task
	if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
		return nil, err
	}

	return task, nil
}
