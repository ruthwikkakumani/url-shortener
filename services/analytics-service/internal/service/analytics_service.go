package service

import (
	"context"
	"time"

	"github.com/ruthwikkakumani/redirection-engine/services/analytics-service/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type AnalyticsService struct {
	repo   *repository.AnalyticsRepo
	logger *zap.Logger
}

func NewAnalyticsService(repo *repository.AnalyticsRepo, logger *zap.Logger) *AnalyticsService {
	return &AnalyticsService{repo: repo, logger: logger}
}

func (s *AnalyticsService) Summary(ctx context.Context, code string) (map[string]any, error) {
	g, gCtx := errgroup.WithContext(ctx)

	var total, unique int64

	g.Go(func() error {
		var err error
		total, err = s.repo.TotalClicks(gCtx, code)
		if err != nil {
			s.logger.Error("summary: total clicks query failed", zap.String("code", code), zap.Error(err))
		}
		return err
	})

	g.Go(func() error {
		var err error
		unique, err = s.repo.UniqueIPs(gCtx, code)
		if err != nil {
			s.logger.Error("summary: unique ips query failed", zap.String("code", code), zap.Error(err))
		}
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return map[string]any{
		"short_code":     code,
		"total_clicks":   total,
		"unique_visitors": unique,
	}, nil
}

func (s *AnalyticsService) OverTime(ctx context.Context, code, interval string) (any, error) {
	since := sinceForInterval(interval)

	switch interval {
	case "hour":
		return s.repo.ClicksOverTimeHourly(ctx, code, since)
	default:
		return s.repo.ClicksOverTimeDaily(ctx, code, since)
	}
}

func sinceForInterval(interval string) time.Time {
	switch interval {
	case "hour":
		return time.Now().UTC().Add(-24 * time.Hour)
	case "week":
		return time.Now().UTC().Add(-7 * 24 * time.Hour)
	default:
		return time.Now().UTC().Add(-30 * 24 * time.Hour)
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
