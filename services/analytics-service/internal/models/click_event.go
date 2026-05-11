package models

import "time"

// ClickEvent mirrors the Kafka message payload published by the redirect-service.
type ClickEvent struct {
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	IP          string    `json:"ip"`
	UserAgent   string    `json:"user_agent"`
	Referer     string    `json:"referer"`
	ClickedAt   time.Time `json:"clicked_at"`
}

// EnrichedClickEvent is the ClickEvent after geo & UA parsing, ready to persist.
type EnrichedClickEvent struct {
	ShortCode   string
	OriginalURL string
	IPAddress   string
	Country     string
	CountryCode string
	City        string
	DeviceType  string // desktop | mobile | tablet
	OS          string
	Browser     string
	Referer     string
	ClickedAt   time.Time
}
