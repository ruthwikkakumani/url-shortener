package service

import (
	"time"

	"github.com/ruthwikkakumani/url-shortener/services/url-service/internal/repository"
	"github.com/ruthwikkakumani/url-shortener/services/url-service/internal/utils"
	"go.uber.org/zap"
)

type UrlService struct {
	logger *zap.Logger
	repo *repository.UrlRepo
}

func NewUrlService(logger *zap.Logger, repo *repository.UrlRepo) (*UrlService) {
	return &UrlService{
		logger: logger,
		repo: repo,
	}
}

func (s *UrlService) CreateShortURL(userId string, url string, expiryMinutes int) (string, error) {
	
	code, err := s.generateUniqueShortCode()
	if err != nil {
		s.logger.Error("unable to generate short code",
			zap.Error(err),
		)
		
		return "", err
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

func (s *UrlService) UpdateOriginalURL(userId string, originalURL *string, code string, expiryMinutes *int) (string, error) {
	
	var expiresAt *time.Time
	if expiryMinutes != nil {
		t := time.Now().Add(time.Duration(*expiryMinutes) * time.Minute)
		expiresAt = &t
	}

	if err := s.repo.UpdateURL(userId, originalURL, code, expiresAt); err != nil {
		s.logger.Error("unable to update data in db",
			zap.Error(err),
		)

		return "", err
	}

	return code, nil
}