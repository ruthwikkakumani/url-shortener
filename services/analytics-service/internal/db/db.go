package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruthwikkakumani/url-shortener/services/analytics-service/internal/config"
	"go.uber.org/zap"
)

type DBService struct {
	logger *zap.Logger
	pool   *pgxpool.Pool
}

func NewDB(logger *zap.Logger) *DBService {
	return &DBService{logger: logger}
}

func (d *DBService) InitDB(ctx context.Context) error {
	if d.pool != nil {
		d.logger.Warn("DB already initialized")
		return nil
	}

	dsn := config.GetEnv("DB_CONN", "")
	if dsn == "" {
		return fmt.Errorf("DB_CONN is not set")
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	cfg.MaxConns = 20
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute

	initCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(initCtx, cfg)
	if err != nil {
		return fmt.Errorf("create pool: %w", err)
	}

	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return fmt.Errorf("ping db: %w", err)
	}

	d.pool = pool
	d.logger.Info("analytics db connected",
		zap.Int32("max_conns", cfg.MaxConns),
		zap.Int32("min_conns", cfg.MinConns),
	)
	return nil
}

func (d *DBService) GetPool() (*pgxpool.Pool, error) {
	if d.pool == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return d.pool, nil
}

func (d *DBService) Close() {
	if d.pool != nil {
		d.logger.Info("closing analytics db connection...")
		d.pool.Close()
	}
}
