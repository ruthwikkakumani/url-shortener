package service

import (
	"time"

	"github.com/ruthwikkakumani/redirection-engine/services/url-service/internal/cache"
	"github.com/ruthwikkakumani/redirection-engine/services/url-service/internal/model"
	"github.com/ruthwikkakumani/redirection-engine/services/url-service/internal/repository"
	"github.com/ruthwikkakumani/redirection-engine/services/url-service/internal/utils"
	"go.uber.org/zap"
	"context"
)

type UrlService struct {
	logger *zap.Logger
	repo   *repository.UrlRepo
	cache  *cache.RedisClient
}

func NewUrlService(logger *zap.Logger, repo *repository.UrlRepo, cache *cache.RedisClient) *UrlService {
	return &UrlService{
		logger: logger,
		repo:   repo,
		cache:  cache,
	}
}

func (s *UrlService) CreateShortURL(userId string, url string, expiryMinutes int, customCode *string) (string, error) {
	var code string
	var err error

	if customCode != nil && *customCode != "" {
		code = *customCode
		if !utils.IsValidShortCode(code) {
			return "", utils.NewError("invalid short code format")
		}
		exists, err := s.repo.ShortCodeExists(code)
		if err != nil {
			return "", err
		}
		if exists {
			return "", utils.NewError("short code already in use")
		}
	} else {
		code, err = s.generateUniqueShortCode()
		if err != nil {
			s.logger.Error("unable to generate short code",
				zap.Error(err),
			)
			return "", err
		}
	}

	var expiresAt *time.Time
	if expiryMinutes > 0 {
		t := time.Now().Add(time.Duration(expiryMinutes) * time.Minute)
		expiresAt = &t
	}

	if err := s.repo.CreateURL(userId, url, code, expiresAt); err != nil {
		s.logger.Error("unable to store data in db",
			zap.Error(err),
		)

		return "", err
	}

	return code, nil
}

func (s *UrlService) generateUniqueShortCode() (string, error) {
	const length = 6

	for {
		code, err := utils.GenerateShortCode(length)
		if err != nil {
			return "", err
		}

		exists, err := s.repo.ShortCodeExists(code)
		if err != nil {
			return "", err
		}

		if !exists {
			return code, nil
		}
	}
}

func (s *UrlService) ListURLS(userId string) ([]model.Url, error) {

	urls, err := s.repo.ListURLS(userId)
	if err != nil {
		s.logger.Error("unable to get data from db",
			zap.Error(err),
		)

		return nil, err
	}

	return urls, nil
}

func (s *UrlService) UpdateURL(userId string, originalURL *string, code string, newCode *string, expiryMinutes *int) (string, error) {

	var finalCode = code
	if newCode != nil && *newCode != "" && *newCode != code {
		if !utils.IsValidShortCode(*newCode) {
			return "", utils.NewError("invalid short code format")
		}
		exists, err := s.repo.ShortCodeExists(*newCode)
		if err != nil {
			return "", err
		}
		if exists {
			return "", utils.NewError("short code already in use")
		}
		finalCode = *newCode
	}

	var expiresAt *time.Time
	if expiryMinutes != nil {
		t := time.Now().Add(time.Duration(*expiryMinutes) * time.Minute)
		expiresAt = &t
	}

	if err := s.repo.UpdateURL(userId, originalURL, code, newCode, expiresAt); err != nil {
		s.logger.Error("unable to update data in db",
			zap.Error(err),
		)

		return "", err
	}

	// Invalidate cache
	if client, err := s.cache.GetClient(); err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		client.Del(ctx, "url:"+code)
		if finalCode != code {
			client.Del(ctx, "url:"+finalCode)
		}
	}

	return finalCode, nil
}

func (s *UrlService) DeleteURL(userId string, shortCode string) error {

	if err := s.repo.DeleteURL(userId, shortCode); err != nil {
		s.logger.Error("unable to delete data in db",
			zap.Error(err),
		)

		return err
	}

	// Invalidate cache
	if client, err := s.cache.GetClient(); err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		client.Del(ctx, "url:"+shortCode)
		s.logger.Info("invalidated cache for deleted url", zap.String("code", shortCode))
	}

	return nil
}
