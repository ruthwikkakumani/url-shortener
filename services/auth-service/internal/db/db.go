package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruthwikkakumani/url-shortener/services/auth-service/internal/config"
	"go.uber.org/zap"
)

type DBService struct{
	logger *zap.Logger
	pool *pgxpool.Pool
}

func NewDB(logger *zap.Logger) (*DBService){
	return &DBService{
		logger: logger,
	}
}

func (db *DBService) InitDB(ctx context.Context) (error) {
	
	if db.pool != nil {
		 db.logger.Warn("DB is already initialized")
			return nil
	}
	
	dbURL := config.GetEnv("DB_CONN_STRING", "")
	
	if dbURL == "" {
		db.logger.Error("database connection string missing", 
			zap.String("env", "DB_CONN_STRING"),
		)
		return fmt.Errorf("database url is not set")
	}
	
	cnfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		db.logger.Error("Failed to parse config", 
			zap.Error(err),
		)
		return fmt.Errorf("parse config: %w", err)
	}
	
	// pool tuning
	cnfg.MaxConns = 20
	cnfg.MinConns = 2
	cnfg.MaxConnLifetime = time.Hour
	cnfg.MaxConnIdleTime = 30 * time.Minute
	
	// Create pool with timeout
	dbCtx, cancel := context.WithTimeout(ctx, 5 * time.Second)
	defer cancel()
	
	pool, err := pgxpool.NewWithConfig(dbCtx, cnfg)
	if err != nil {
		db.logger.Error("Failed to create pool",
			zap.Error(err),
		)
		return fmt.Errorf("create pool: %w", err)
	}
	
	// ping with fresh context
	pingCtx, pingCancel := context.WithTimeout(ctx, 5 * time.Second)
	defer pingCancel()
	
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		db.logger.Error("failed to ping db",
			zap.Error(err),
		)
		return fmt.Errorf("ping db : %w", err)
	}
	
	db.pool = pool
	
	db.logger.Info("database connected successfully",
		zap.Int32("max_conns", cnfg.MaxConns),
		zap.Int32("min_conns", cnfg.MinConns),
		zap.Duration("max_conn_lifetime", cnfg.MaxConnLifetime),
		zap.Duration("max_conn_idle_time", cnfg.MaxConnIdleTime),
	)
	
	return nil
}

func (db *DBService) Close() {
	if db.pool != nil{
		db.logger.Info("closing database connection...")
		db.pool.Close()
	}
}

func (db *DBService) GetPool() (*pgxpool.Pool, error){
	if db.pool == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return db.pool, nil
}