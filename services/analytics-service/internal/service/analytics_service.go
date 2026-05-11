package service

import (
	"context"
	"time"

	"github.com/ruthwikkakumani/url-shortener/services/analytics-service/internal/repository"
	"go.uber.org/zap"
)

type AnalyticsService struct {
	repo   *repository.AnalyticsRepo
	logger *zap.Logger
}

func NewAnalyticsService(repo *repository.AnalyticsRepo, logger *zap.Logger) *AnalyticsService {
	return &AnalyticsService{repo: repo, logger: logger}
}

// Summary returns total clicks and unique visitors for a short code.
func (s *AnalyticsService) Summary(ctx context.Context, code string) (map[string]any, error) {
	total, err := s.repo.TotalClicks(ctx, code)
	if err != nil {
		s.logger.Error("summary: total clicks query failed", zap.String("code", code), zap.Error(err))
		return nil, err
	}

	unique, err := s.repo.UniqueIPs(ctx, code)
	if err != nil {
		s.logger.Error("summary: unique ips query failed", zap.String("code", code), zap.Error(err))
		return nil, err
	}

	return map[string]any{
		"short_code":     code,
		"total_clicks":   total,
		"unique_visitors": unique,
	}, nil
}

// OverTime returns hourly or daily time-series based on the interval param.
func (s *AnalyticsService) OverTime(ctx context.Context, code, interval string) (any, error) {
	since := sinceForInterval(interval)

	switch interval {
	case "hour":
		return s.repo.ClicksOverTimeHourly(ctx, code, since)
	default: // "day" or anything else → daily
		return s.repo.ClicksOverTimeDaily(ctx, code, since)
	}
}

func sinceForInterval(interval string) time.Time {
	switch interval {
	case "hour":
		return time.Now().UTC().Add(-24 * time.Hour) // last 24 hours, hourly
	case "week":
		return time.Now().UTC().Add(-7 * 24 * time.Hour) // last 7 days, daily
	default: // "day"
		return time.Now().UTC().Add(-30 * 24 * time.Hour) // last 30 days, daily
	}
}

func (s *AnalyticsService) Countries(ctx context.Context, code string) (any, error) {
	return s.repo.ByCountry(ctx, code)
}

func (s *AnalyticsService) Cities(ctx context.Context, code string) (any, error) {
	return s.repo.ByCity(ctx, code)
}

func (s *AnalyticsService) Devices(ctx context.Context, code string) (any, error) {
	return s.repo.ByDevice(ctx, code)
}

func (s *AnalyticsService) OSBreakdown(ctx context.Context, code string) (any, error) {
	return s.repo.ByOS(ctx, code)
}

func (s *AnalyticsService) Browsers(ctx context.Context, code string) (any, error) {
	return s.repo.ByBrowser(ctx, code)
}

func (s *AnalyticsService) PeakHours(ctx context.Context, code string) (any, error) {
	return s.repo.PeakHours(ctx, code)
}

func (s *AnalyticsService) RecentClicks(ctx context.Context, code string, limit int) (any, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.RecentClicks(ctx, code, limit)
}
