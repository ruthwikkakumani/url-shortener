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