package repository

import (
	"context"
	"time"

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

func (r *UrlRepo) ShortCodeExists(code string) (bool, error) {
	var exists bool
	
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE short_code=$1)`
	err := r.db.QueryRow(context.Background(), query, code).Scan(&exists)
	
	return exists, err
}

func (r *UrlRepo) CreateURL(userID, originalURL, code string, expiresAt *time.Time) error {
	query := `
		INSERT INTO urls (user_id, original_url, short_code, expires_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(context.Background(), query, userID, originalURL, code, expiresAt)
	return err
}

func (r *UrlRepo) GetOriginalURL(code string) (string, error) {
	var originalURL string
	query := `SELECT original_url FROM urls WHERE short_code=$1 AND (expires_at IS NULL OR expires_at > NOW())`
	err := r.db.QueryRow(context.Background(), query, code).Scan(&originalURL)
	return originalURL, err
}