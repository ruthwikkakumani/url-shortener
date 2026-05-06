package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ruthwikkakumani/url-shortener/services/url-service/internal/config"
	"github.com/ruthwikkakumani/url-shortener/services/url-service/internal/service"
	"go.uber.org/zap"
)

type UrlHandler struct {
	logger  *zap.Logger
	service *service.UrlService
}

type ShortenURLRequest struct {
	OriginalURL   string `json:"original_url" binding:"required,url"`
	ExpiryMinutes int    `json:"expiry_minutes" binding:"omitempty,gte=1,lte=10080"`
}

type UpdateURLRequest struct {
	OriginalURL   *string `json:"original_url" binding:"omitempty,url"`
	ExpiryMinutes *int    `json:"expiry_minutes" binding:"omitempty,gte=1,lte=10080"`
}

// URL initializer
func NewUrlHandler(logger *zap.Logger, service *service.UrlService) *UrlHandler {
	return &UrlHandler{
		logger:  logger,
		service: service,
	}
}

func (h *UrlHandler) ShortenURL(c *gin.Context) {

	userId := c.GetHeader("X-User-ID")
	if userId == "" {
		h.logger.Error("user not authenticated")

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	var req ShortenURLRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid shorten request",
			zap.Error(err),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	code, err := h.service.CreateShortURL(userId, req.OriginalURL, req.ExpiryMinutes)
	if err != nil {
		h.logger.Error("failed to create short url", zap.Error(err))

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	base := config.GetEnv("BASE_URL", "http://localhost:8082")

	c.JSON(http.StatusOK, gin.H{
		"short_url": base + "/r/" + code,
	})

}

func (h *UrlHandler) UpdateURL(c *gin.Context) {

	userId := c.GetHeader("X-User-ID")
	if userId == "" {
		h.logger.Error("user not authenticated")

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	var req UpdateURLRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid update request",
			zap.Error(err),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	if req.OriginalURL == nil &&
		req.ExpiryMinutes == nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "at least one field is required",
		})
		return
	}

	code := c.Param("shortCode")

	code, err := h.service.UpdateOriginalURL(userId, req.OriginalURL, code, req.ExpiryMinutes)
	if err != nil {
		h.logger.Error("failed to update url", zap.Error(err))

		if err.Error() == "url not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "url not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	base := config.GetEnv("BASE_URL", "http://localhost:8082")

	c.JSON(http.StatusOK, gin.H{
		"short_url": base + "/r/" + code,
	})
}
