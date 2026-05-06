package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

func (r *UrlRepo) CreateURL(userID, originalURL, code string, expiresAt *time.Time) (error) {
	query := `
		INSERT INTO urls (user_id, original_url, short_code, expires_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(context.Background(), query, userID, originalURL, code, expiresAt)
	return err
}

func (r *UrlRepo) UpdateURL(userID string, originalURL *string, code string, expiresAt *time.Time) (error) {

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
		context.Background(),
		query,
		args...,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("url not found")
	}

	return nil
}