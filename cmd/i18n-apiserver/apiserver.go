package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/xmualex2023/i18n-translation/internal/apiserver/config"
	"github.com/xmualex2023/i18n-translation/internal/apiserver/controller"
	"github.com/xmualex2023/i18n-translation/internal/apiserver/repository"
	"github.com/xmualex2023/i18n-translation/internal/apiserver/service"
	"github.com/xmualex2023/i18n-translation/internal/pkg/auth"
	"github.com/xmualex2023/i18n-translation/internal/pkg/limiter"
	"github.com/xmualex2023/i18n-translation/internal/pkg/llm"
	"github.com/xmualex2023/i18n-translation/internal/pkg/metrics"
	"github.com/xmualex2023/i18n-translation/internal/pkg/middleware"
	"github.com/xmualex2023/i18n-translation/internal/pkg/queue"
	"github.com/xmualex2023/i18n-translation/internal/pkg/util"
	"github.com/xmualex2023/i18n-translation/internal/pkg/worker"
)

var (
	configPath = flag.String("config", "configs/apiserver.yaml", "配置文件路径")
)

func main() {
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	gin.SetMode(cfg.Server.Mode)

	repo, err := repository.NewRepository(cfg)
	if err != nil {
		log.Fatalf("failed to initialize storage layer: %v", err)
	}

	llmClient := llm.NewClient(cfg.LLM.APIKey, cfg.LLM.Endpoint)

	// initialize redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// initialize task queue
	taskQueue := queue.NewRedisQueue(redisClient, "translation_tasks")

	tokenCache := auth.NewRedisTokenCache(redisClient, "token", cfg.JWT.Expire)
	// initialize service layer
	svc := service.NewService(cfg, repo, llmClient, taskQueue, tokenCache)

	// initialize worker
	worker := worker.NewWorker(taskQueue, svc.HandleTranslationTask)
	worker.Start(cfg.Worker.Count) // 启动指定数量的工作器
	defer worker.Stop()

	// initialize controller
	ctrl := controller.NewController(svc)

	// create router
	router := setupRouter(ctrl, cfg)

	// create http server
	srv := &http.Server{
		Addr:    cfg.Server.HTTP.Address,
		Handler: router,
	}

	// graceful shutdown
	util.SafetyGo(func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to listen: %v", err)
		}
	})

	// wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server force shutdown:", err)
	}

	log.Println("server exited")
}

func setupRouter(ctrl controller.IController, cfg *config.Config) *gin.Engine {
	r := gin.New()

	// install middleware
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/metrics"},
		Formatter: func(param gin.LogFormatterParams) string {
			keys := param.Keys
			userID, ok := keys["user_id"]
			if !ok {
				userID = "unknown"
			}
			return fmt.Sprintf("[%v] UserID: %s - %s %s - %13v - %3d - %s\n%s",
				param.TimeStamp.Format("2006-01-02 15:04:05"),
				userID,
				param.Method,
				param.Path,
				param.Latency,
				param.StatusCode,
				param.ClientIP,
				param.ErrorMessage,
			)
		},
	}))
	r.Use(gin.Recovery())
	r.Use(middleware.ErrorHandler())

	prom, err := middleware.NewPrometheus(cfg)
	if err != nil {
		log.Fatalf("failed to initialize prometheus: %v", err)
	}
	prom.RegisterCollector(metrics.CollectorVector...)
	r.Use(prom.MetricsMiddleware())
	util.SafetyGo(prom.Run)

	// initialize redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// initialize rate limiter
	rateLimiter := limiter.NewRateLimiter(
		redisClient,
		cfg.RateLimit.MaxRequests,
		cfg.RateLimit.Duration,
	)

	// initialize token cache
	tokenCache := auth.NewRedisTokenCache(
		redisClient,
		"token",
		cfg.JWT.Expire,
	)

	// create jwt maker
	jwtMaker := auth.NewJWTMaker(cfg.JWT.Secret, tokenCache)

	// api group
	api := r.Group("/api/v1")
	api.Use(middleware.RateLimiter(rateLimiter))
	{
		// public routes
		api.POST("/auth/register", ctrl.Register)
		api.POST("/auth/login", ctrl.Login)
		api.POST("/auth/refresh", ctrl.RefreshToken)

		// authorized routes
		authorized := api.Group("/tasks")
		authorized.Use(middleware.AuthMiddleware(jwtMaker))
		{
			authorized.POST("", ctrl.CreateTask)
			authorized.POST("/:taskID/translate", ctrl.ExecuteTranslation)
			authorized.GET("/:taskID", ctrl.GetTaskStatus)
			authorized.GET("/:taskID/download", ctrl.DownloadTranslation)
		}
	}

	// run pprof
	util.SafetyGo(func() {
		runProfile(cfg)
	})

	return r
}

func runProfile(cfg *config.Config) {
	err := http.ListenAndServe(cfg.Pprof.Address, nil)
	if err != nil {
		log.Fatalf("failed to run profile: %v", err)
	}
}
