package service

import (
	"context"

	"github.com/xmualex2023/i18n-translation/internal/apiserver/config"
	"github.com/xmualex2023/i18n-translation/internal/apiserver/repository"
	"github.com/xmualex2023/i18n-translation/internal/pkg/auth"
	"github.com/xmualex2023/i18n-translation/internal/pkg/queue"
)

type translator interface {
	Translate(ctx context.Context, text, sourceLang, targetLang string) (string, error)
}

type Service struct {
	cfg        *config.Config
	repo       *repository.Repository
	translator translator
	queue      queue.Queue
	cache      auth.TokenCache
}

func NewService(cfg *config.Config, repo *repository.Repository, tr translator, q queue.Queue, cache auth.TokenCache) *Service {
	return &Service{
		cfg:        cfg,
		repo:       repo,
		translator: tr,
		queue:      q,
		cache:      cache,
	}
}
