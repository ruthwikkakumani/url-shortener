package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/ruthwikkakumani/redirection-engine/services/url-service/internal/config"
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

	host := config.GetEnv("REDIS_HOST", "")
	port := config.GetEnv("REDIS_PORT", "")
	password := config.GetEnv("REDIS_PASSWORD", "")
	username := config.GetEnv("REDIS_USER", "")

	if host == "" || port == "" {
		return fmt.Errorf("REDIS_HOST or REDIS_PORT not set")
	}

	addr := fmt.Sprintf("%s:%s", host, port)

	opt := &redis.Options{
		Addr:            addr,
		Username:        username,
		Password:        password,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		PoolSize:        50,
		MinIdleConns:    10,
		MaxRetries:      3,
		MinRetryBackoff: 100 * time.Millisecond,
		MaxRetryBackoff: 1 * time.Second,
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
		if err := r.client.Close(); err != nil {
			r.logger.Error("failed to close redis", zap.Error(err))
		}
	}
}

func (r *RedisClient) GetClient() (*redis.Client, error) {
	if r.client == nil {
		return nil, fmt.Errorf("redis not initialized")
	}

	return r.client, nil
}