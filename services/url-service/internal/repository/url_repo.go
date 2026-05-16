package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruthwikkakumani/redirection-engine/services/url-service/internal/model"
	"go.uber.org/zap"
)

type UrlRepo struct {
	logger *zap.Logger
	db     *pgxpool.Pool
}

func NewUrlRepo(logger *zap.Logger, db *pgxpool.Pool) *UrlRepo {
	return &UrlRepo{
		logger: logger,
		db:     db,
	}
}

func (r *UrlRepo) ShortCodeExists(ctx context.Context, code string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE short_code=$1)`
	err := r.db.QueryRow(ctx, query, code).Scan(&exists)
	return exists, err
}

func (r *UrlRepo) CreateURL(ctx context.Context, userID, originalURL, code string, expiresAt *time.Time) error {
	query := `
		INSERT INTO urls (user_id, original_url, short_code, expires_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.Exec(ctx, query, userID, originalURL, code, expiresAt)
	return err
}

func (r *UrlRepo) ListURLS(ctx context.Context, userID string) ([]model.Url, error) {
	query := `
		SELECT short_code, original_url, created_at, expires_at FROM urls
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	var urls []model.Url

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return urls, err
	}
	defer rows.Close()

	for rows.Next() {
		var url model.Url

		err := rows.Scan(
			&url.ShortCode,
			&url.OriginalURL,
			&url.CreatedAt,
			&url.ExpiresAt,
		)

		if err != nil {
			return nil, err
		}

		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func (r *UrlRepo) UpdateURL(ctx context.Context, userID string, originalURL *string, code string, newCode *string, expiresAt *time.Time) error {
	setClauses := []string{}
	args := []interface{}{}
	argPos := 1

	if originalURL != nil {
		setClauses = append(
			setClauses,
			fmt.Sprintf("original_url = $%d", argPos),
		)
		args = append(args, *originalURL)
		argPos++
	}

	if newCode != nil {
		setClauses = append(
			setClauses,
			fmt.Sprintf("short_code = $%d", argPos),
		)
		args = append(args, *newCode)
		argPos++
	}

	if expiresAt != nil {
		setClauses = append(
			setClauses,
			fmt.Sprintf("expires_at = $%d", argPos),
		)
		args = append(args, *expiresAt)
		argPos++
	}

	if len(setClauses) == 0 {
		return errors.New("no fields to update")
	}

	query := fmt.Sprintf(`
		UPDATE urls
		SET %s
		WHERE user_id = $%d
		AND short_code = $%d
	`,
		strings.Join(setClauses, ", "),
		argPos,
		argPos+1,
	)

	args = append(args, userID, code)

	result, err := r.db.Exec(
		ctx, 
		query, 
		args...
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("url not found")
	}

	return nil
}

func (r *UrlRepo) DeleteURL(ctx context.Context, userID string, shortCode string) error {
	query := `
		DELETE FROM urls
		WHERE user_id = $1
  		AND short_code = $2;
	`

	result, err := r.db.Exec(ctx, query, userID, shortCode)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("url not found")
	}

	return nil
}