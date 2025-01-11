package middleware

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/xmualex2023/i18n-translation/internal/apiserver/config"
)

var (
	defaultPushInterval     = 2
	defaultHistogramBuckets = []float64{0.001, 0.002, 0.005, 0.01, 0.02, 0.05, 0.1, 0.5, 0.8, 1, 3}
)

type Prometheus struct {
	cfg       *config.Config
	gather    prometheus.Gatherer
	registry  prometheus.Registerer
	close     chan struct{}
	histogram *prometheus.HistogramVec
}

func NewPrometheus(cfg *config.Config) (*Prometheus, error) {
	if cfg.Metrics.PullHost == "" && cfg.Metrics.URL == "" {
		return nil, fmt.Errorf("prometheus config error: pullHost and url can not be empty at the same time")
	}
	if cfg.Metrics.Job == "" {
		cfg.Metrics.Job = filepath.Base(os.Args[0])
	}
	if cfg.Metrics.Instance == "" {
		instance, err := os.Hostname()
		if err != nil {
			return nil, err
		}
		cfg.Metrics.Instance = instance
	}
	if cfg.Metrics.PushIntervalSec == 0 {
		cfg.Metrics.PushIntervalSec = defaultPushInterval
	}
	if len(cfg.Metrics.Buckets) == 0 {
		cfg.Metrics.Buckets = defaultHistogramBuckets
	}
	register := prometheus.NewRegistry()
	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_sec_histogram",
		Help:    "histogram of http API time cost",
		Buckets: cfg.Metrics.Buckets,
	}, []string{"status", "method", "pattern", "service"})
	register.MustRegister(histogram)
	register.MustRegister(collectors.NewGoCollector(),
		collectors.NewProcessCollector(
			collectors.ProcessCollectorOpts{Namespace: cfg.Metrics.Job}))
	return &Prometheus{
		cfg:       cfg,
		gather:    register,
		registry:  register,
		close:     make(chan struct{}),
		histogram: histogram,
	}, nil
}

func (prom *Prometheus) RegisterCollector(cols ...prometheus.Collector) {
	for _, c := range cols {
		prom.registry.MustRegister(c)
	}
}

func (prom *Prometheus) Register() prometheus.Registerer {
	return prom.registry
}

func (prom *Prometheus) Gatherer() prometheus.Gatherer {
	return prom.gather
}

// TODO: 需要使用日志
func (prom *Prometheus) Run() {
	// pull模式
	if prom.cfg.Metrics.PullHost != "" {
		http.Handle("/metrics", promhttp.HandlerFor(prom.gather, promhttp.HandlerOpts{
			// ErrorLog: log.Std,
		}))
		// log.Info("prom http is running at", prom.cfg.PullHost)
		// log.Fatal(http.ListenAndServe(prom.cfg.PullHost, nil))
	}
	// 没有开启prometheus
	if prom.cfg.Metrics.URL == "" {
		return
	}
	// push 模式
	ticker := time.NewTicker(time.Duration(prom.cfg.Metrics.PushIntervalSec) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// log.Debug("pushing prometheus...")
			err := push.New(prom.cfg.Metrics.URL, prom.cfg.Metrics.Job).Grouping("instance", prom.cfg.Metrics.Instance).Gatherer(prom.gather).Add()
			if err != nil {
				// log.Error("prometheus.Push failed:", err)
			}
		case <-prom.close:
			// log.Info("closed")
			return
		}
	}
}

func (prom *Prometheus) Close() {
	close(prom.close)
}

func (prom *Prometheus) MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		prom.histogram.With(prometheus.Labels{"service": prom.cfg.Metrics.Job, "method": c.Request.Method, "status": fmt.Sprintf("%d", c.Writer.Status()),
			"pattern": c.FullPath(),
		}).Observe(float64(time.Since(start).Seconds()))
	}
}
