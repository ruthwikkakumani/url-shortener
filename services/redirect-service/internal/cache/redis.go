package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/ruthwikkakumani/url-shortener/services/redirect-service/internal/config"
	"go.uber.org/zap"
)

type RedisClient struct {
	logger *zap.Logger
	client *redis.Client

}

func NewRedisClient(logger *zap.Logger) (*RedisClient) {
	return &RedisClient{
		logger: logger,
	}
}

func (r *RedisClient) Init(ctx context.Context) (error) {
	if r.client != nil {
		r.logger.Warn("Redis already initialized")
		return nil
	}

	addr := config.GetEnv("REDIS_ADDR","")

	opt := &redis.Options{
		Addr: addr,
		DialTimeout: 5 * time.Second,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second,
		PoolSize: 10,
	}

	client := redis.NewClient(opt)

	pingCtx, cancel := context.WithTimeout(ctx, 5 * time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		r.logger.Error("failed to connect redis",
			zap.Error(err),
		)
		return fmt.Errorf("redis ping failed: %w", err)
	}

	r.client = client

	r.logger.Info("redis connected successfully",
		zap.String("addr", addr),
		zap.Int("pool_size", opt.PoolSize),
	)

	return nil
}

func (r *RedisClient) Close() {
	if r.client != nil {
		r.logger.Info("closing redis connection...")
		_ = r.client.Close()
	}
}

func (r *RedisClient) GetClient() (*redis.Client, error) {
	if r.client == nil {
		return nil, fmt.Errorf("redis not initialized")
	}

	return r.client, nil
}