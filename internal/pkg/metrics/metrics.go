package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// TaskCounter 任务计数器
	TaskCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "translation_tasks_total",
			Help: "翻译任务总数",
		},
		[]string{"status"}, // pending, processing, completed, failed
	)

	// TaskDuration 任务处理时间
	TaskDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "translation_task_duration_seconds",
			Help:    "翻译任务处理时间",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"},
	)

	// QueueSize 队列大小
	QueueSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "translation_queue_size",
			Help: "翻译任务队列大小",
		},
	)

	// WorkerCount 工作器数量
	WorkerCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "translation_worker_count",
			Help: "活跃的工作器数量",
		},
	)

	// TODO: Add more metrics here

	// append metrics
	CollectorVector = []prometheus.Collector{
		TaskCounter,
		TaskDuration,
		QueueSize,
		WorkerCount,
	}
)

func IncTaskCounter(status string) {
	TaskCounter.WithLabelValues(status).Inc()
}

func ObserveTaskDuration(status string, duration time.Duration) {
	TaskDuration.WithLabelValues(status).Observe(duration.Seconds())
}

func SetQueueSize(size int) {
	QueueSize.Set(float64(size))
}

func SetWorkerCount(count int) {
	WorkerCount.Set(float64(count))
}
