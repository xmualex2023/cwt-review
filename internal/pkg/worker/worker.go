package worker

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xmualex2023/i18n-translation/internal/pkg/metrics"
	"github.com/xmualex2023/i18n-translation/internal/pkg/queue"
)

// Handler 任务处理函数
type Handler func(context.Context, queue.Task) error

// Worker 工作器
type Worker struct {
	queue      queue.Queue
	handler    Handler
	stopChan   chan struct{}
	wg         sync.WaitGroup
	activeJobs int32 // 活跃任务数
}

func NewWorker(queue queue.Queue, handler Handler) *Worker {
	return &Worker{
		queue:    queue,
		handler:  handler,
		stopChan: make(chan struct{}),
	}
}

// Start 启动工作器
func (w *Worker) Start(workerCount int) {
	for i := 0; i < workerCount; i++ {
		w.wg.Add(1)
		go w.run()
	}
}

// Stop 停止工作器
func (w *Worker) Stop() {
	close(w.stopChan)
	w.wg.Wait()
}

func (w *Worker) run() {
	defer w.wg.Done()

	for {
		select {
		case <-w.stopChan:
			return
		default:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			task, err := w.queue.Dequeue(ctx)
			cancel()

			if err != nil {
				log.Printf("从队列获取任务失败: %v", err)
				continue
			}

			// 更新活跃任务数
			atomic.AddInt32(&w.activeJobs, 1)
			metrics.SetWorkerCount(int(atomic.LoadInt32(&w.activeJobs)))

			start := time.Now()
			err = w.handler(context.Background(), task)
			duration := time.Since(start)

			// 更新指标
			status := "completed"
			if err != nil {
				status = "failed"
				log.Printf("处理任务失败: %v", err)
			}
			metrics.IncTaskCounter(status)
			metrics.ObserveTaskDuration(status, duration)

			// 减少活跃任务数
			atomic.AddInt32(&w.activeJobs, -1)
			metrics.SetWorkerCount(int(atomic.LoadInt32(&w.activeJobs)))
		}
	}
}
