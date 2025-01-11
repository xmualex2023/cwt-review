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

// Handler task handler function
type Handler func(context.Context, queue.Task) error

// Worker
type Worker struct {
	queue      queue.Queue
	handler    Handler
	stopChan   chan struct{}
	wg         sync.WaitGroup
	activeJobs int32
}

func NewWorker(queue queue.Queue, handler Handler) *Worker {
	return &Worker{
		queue:    queue,
		handler:  handler,
		stopChan: make(chan struct{}),
	}
}

// Start start worker
func (w *Worker) Start(workerCount int) {
	for i := 0; i < workerCount; i++ {
		w.wg.Add(1)
		go w.run()
	}
}

// Stop stop worker
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
				log.Printf("failed to dequeue task, error: %v", err)
				continue
			}

			// update active jobs
			atomic.AddInt32(&w.activeJobs, 1)
			metrics.SetWorkerCount(int(atomic.LoadInt32(&w.activeJobs)))

			start := time.Now()
			err = w.handler(context.Background(), task)
			duration := time.Since(start)

			// update metrics
			status := "completed"
			if err != nil {
				status = "failed"
				log.Printf("failed to handle task, error: %v", err)
			}
			metrics.IncTaskCounter(status)
			metrics.ObserveTaskDuration(status, duration)

			// decrease active jobs
			atomic.AddInt32(&w.activeJobs, -1)
			metrics.SetWorkerCount(int(atomic.LoadInt32(&w.activeJobs)))
		}
	}
}
