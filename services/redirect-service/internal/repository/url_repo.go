package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type UrlRepo struct {
	logger *zap.Logger
	db *pgxpool.Pool
}

func NewUrlRepo(logger *zap.Logger, db *pgxpool.Pool) (*UrlRepo){
	return &UrlRepo{
		logger: logger,
		db: db,
	}
}

func (r *UrlRepo) GetOriginalURL(code string) (string, error) {
	var originalURL string
	query := `SELECT original_url FROM urls WHERE short_code=$1 AND (expires_at IS NULL OR expires_at > NOW())`
	err := r.db.QueryRow(context.Background(), query, code).Scan(&originalURL)
	return originalURL, err
}