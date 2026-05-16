package service

import (
	"context"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/ruthwikkakumani/redirection-engine/services/redirect-service/internal/cache"
	"github.com/ruthwikkakumani/redirection-engine/services/redirect-service/internal/repository"
	"go.uber.org/zap"
)

type UrlService struct {
	logger *zap.Logger
	repo   *repository.UrlRepo
	cache  *cache.RedisClient
}

func NewUrlService(logger *zap.Logger, repo *repository.UrlRepo, cache *cache.RedisClient) *UrlService {
	return &UrlService{
		logger: logger,
		repo:   repo,
		cache:  cache,
	}
}

func (s *UrlService) GetOriginalURL(ctx context.Context, code string) (string, error) {

	ctx, cancel := context.WithTimeout(ctx, 2 * time.Second)
	defer cancel()

	client, err := s.cache.GetClient()
	if err != nil {
		s.logger.Warn("Redis unavailable, fallback to db",
			zap.Error(err),
		)
		return s.repo.GetOriginalURL(code)
	}

	key := "url:" + code
	val, err := client.Get(ctx, key).Result()
	if err == nil {
		return val, nil
	}

	if err != redis.Nil {
	    s.logger.Warn("redis error",
	        zap.String("code", code),
	        zap.Error(err),
	    )
	}


	url, err := s.repo.GetOriginalURL(code)
	if err != nil {
		s.logger.Error("failed to get original url", 
			zap.String("code", code), 
			zap.Error(err),
		)
		return "", err
	}

	go func() {
		warmCtx, warmCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer warmCancel()
		ttl := 10*time.Minute + time.Duration(rand.Intn(60))*time.Second
		if err := client.Set(warmCtx, key, url, ttl).Err(); err != nil {
			s.logger.Warn("Failed to set cache", zap.Error(err))
		}
	}()

	return url, nil
}