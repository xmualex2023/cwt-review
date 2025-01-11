package queue

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTask 用于测试的任务实现
type MockTask struct {
	ID string `json:"id"`
}

func (t *MockTask) GetID() string {
	return t.ID
}

// setupTestRedis 创建测试用的 Redis 实例
func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, func() {
		client.Close()
		mr.Close()
	}
}

func TestRedisQueue(t *testing.T) {
	// 测试用例组
	tests := []struct {
		name     string
		testFunc func(*testing.T, *RedisQueue)
	}{
		{"TestEnqueueDequeue", testEnqueueDequeue},
		{"TestEmptyDequeue", testEmptyDequeue},
		{"TestMultipleEnqueueDequeue", testMultipleEnqueueDequeue},
		{"TestConcurrentEnqueueDequeue", testConcurrentEnqueueDequeue},
	}

	// 运行所有测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := setupTestRedis(t)
			defer cleanup()

			queue := NewRedisQueue(client, "test_queue")
			tt.testFunc(t, queue)
		})
	}
}

// 测试基本的入队和出队功能
func testEnqueueDequeue(t *testing.T, q *RedisQueue) {
	ctx := context.Background()
	task := &MockTask{ID: "task1"}

	// 测试入队
	err := q.Enqueue(ctx, task)
	require.NoError(t, err)

	// 测试出队
	receivedTask, err := q.Dequeue(ctx)
	require.NoError(t, err)

	// 验证任务内容
	mockTask, ok := receivedTask.(*MockTask)
	require.True(t, ok)
	assert.Equal(t, task.ID, mockTask.ID)
}

// 测试空队列出队
func testEmptyDequeue(t *testing.T, q *RedisQueue) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 从空队列出队应该超时
	_, err := q.Dequeue(ctx)
	assert.Error(t, err)
}

// 测试多个任务的入队和出队
func testMultipleEnqueueDequeue(t *testing.T, q *RedisQueue) {
	ctx := context.Background()
	tasks := []*MockTask{
		{ID: "task1"},
		{ID: "task2"},
		{ID: "task3"},
	}

	// 入队多个任务
	for _, task := range tasks {
		err := q.Enqueue(ctx, task)
		require.NoError(t, err)
	}

	// 验证出队顺序
	for _, expectedTask := range tasks {
		receivedTask, err := q.Dequeue(ctx)
		require.NoError(t, err)

		mockTask, ok := receivedTask.(*MockTask)
		require.True(t, ok)
		assert.Equal(t, expectedTask.ID, mockTask.ID)
	}
}

// 测试并发入队和出队
func testConcurrentEnqueueDequeue(t *testing.T, q *RedisQueue) {
	ctx := context.Background()
	taskCount := 100
	done := make(chan bool)

	// 并发入队
	go func() {
		for i := 0; i < taskCount; i++ {
			task := &MockTask{ID: fmt.Sprintf("task%d", i)}
			err := q.Enqueue(ctx, task)
			require.NoError(t, err)
		}
		done <- true
	}()

	// 并发出队
	receivedCount := 0
	go func() {
		for receivedCount < taskCount {
			_, err := q.Dequeue(ctx)
			require.NoError(t, err)
			receivedCount++
		}
		done <- true
	}()

	// 等待所有操作完成
	<-done
	<-done

	assert.Equal(t, taskCount, receivedCount)
}

// TestNewRedisQueue 测试队列初始化
func TestNewRedisQueue(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	queue := NewRedisQueue(client, "test_queue")
	assert.NotNil(t, queue)
	assert.Equal(t, "test_queue", queue.queueKey)
	assert.Equal(t, 3, queue.retryCount)
	assert.Equal(t, 5*time.Second, queue.retryDelay)
}

// TestQueueSize 测试队列大小统计
func TestQueueSize(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	queue := NewRedisQueue(client, "test_queue")
	ctx := context.Background()

	// 入队多个任务
	taskCount := 5
	for i := 0; i < taskCount; i++ {
		task := &MockTask{ID: fmt.Sprintf("task%d", i)}
		err := queue.Enqueue(ctx, task)
		require.NoError(t, err)
	}

	// 验证队列大小
	size, err := client.LLen(ctx, queue.queueKey).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(taskCount), size)

	// 出队并验证队列大小变化
	for i := taskCount; i > 0; i-- {
		_, err := queue.Dequeue(ctx)
		require.NoError(t, err)

		size, err := client.LLen(ctx, queue.queueKey).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(i-1), size)
	}
}
