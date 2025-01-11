package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xmualex2023/i18n-translation/internal/pkg/metrics"
)

// Task task interface
type Task interface {
	GetID() string
}

// Queue queue interface
type Queue interface {
	Enqueue(ctx context.Context, task Task) error
	Dequeue(ctx context.Context) (Task, error)
}

// RedisQueue redis queue implementation
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

// Enqueue add task to queue
func (q *RedisQueue) Enqueue(ctx context.Context, task Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task, error: %w", err)
	}

	if err := q.client.LPush(ctx, q.queueKey, data).Err(); err != nil {
		return fmt.Errorf("failed to enqueue task, error: %w", err)
	}

	// update queue size
	size, err := q.client.LLen(ctx, q.queueKey).Result()
	if err == nil {
		metrics.SetQueueSize(int(size))
	}

	return nil
}

// Dequeue get task from queue
func (q *RedisQueue) Dequeue(ctx context.Context) (Task, error) {
	result, err := q.client.BRPop(ctx, 0, q.queueKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to dequeue task, error: %w", err)
	}

	// update queue size
	size, err := q.client.LLen(ctx, q.queueKey).Result()
	if err == nil {
		metrics.SetQueueSize(int(size))
	}

	var task Task
	if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task, error: %w", err)
	}

	return task, nil
}
