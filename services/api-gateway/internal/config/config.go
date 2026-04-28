package config

import (
	"net/url"
	"os"

	"go.uber.org/zap"
)

type Config struct {
	Port           string
	JWTSecret      string
	AuthServiceURL string
	URLServiceURL  string
	AnalyticsURL   string
}

func Load(logger *zap.Logger) *Config {
	cfg := &Config{
		Port:           GetEnv("PORT", "8000"),
		JWTSecret:      GetEnv("JWT_SECRET", ""),
		AuthServiceURL: GetEnv("AUTH_SERVICE_URL", ""),
		URLServiceURL:  GetEnv("URL_SERVICE_URL", ""),
		AnalyticsURL:   GetEnv("ANALYTICS_SERVICE_URL", ""),
	}

	must(cfg.JWTSecret, "JWT_SECRET", logger)
	must(cfg.AuthServiceURL, "AUTH_SERVICE_URL", logger)
	must(cfg.URLServiceURL, "URL_SERVICE_URL", logger)

	// Validate URL format
	validateURL(cfg.AuthServiceURL, "AUTH_SERVICE_URL", logger)
	validateURL(cfg.URLServiceURL, "URL_SERVICE_URL", logger)

	if cfg.AnalyticsURL == "" {
		logger.Warn("ANALYTICS_SERVICE_URL not set (analytics disabled)")
	}

	return cfg
}

func must(val, name string, logger *zap.Logger) {
	if val == "" {
		logger.Fatal(name + " is required")
	}
}

func validateURL(raw, name string, logger *zap.Logger) {
	u, err := url.ParseRequestURI(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		logger.Fatal("invalid " + name + ": " + raw)
	}
}

func GetEnv(key string, defaultValue string) (string) {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	return value
}