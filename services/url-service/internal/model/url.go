package model

import "time"

type Url struct {
	ID string `json:"-" db:"id"`
	UserID string `json:"-" db:"user_id"`
	ShortCode string `json:"short_code" db:"short_code"`
	OriginalURL string `json:"original_url" db:"original_url"`
	CreatedAt time.Time  `json:"-" db:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"`
}