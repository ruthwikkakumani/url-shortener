package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ruthwikkakumani/url-shortener/services/url-service/internal/config"
	"github.com/ruthwikkakumani/url-shortener/services/url-service/internal/service"
	"go.uber.org/zap"
)

type UrlHandler struct {
	logger *zap.Logger
	service *service.UrlService
}

type ShortenURLRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
	ExpiryMinutes int `json:"expiry_minutes" binding:"omitempty,gte=1,lte=10080"`
}

// URL initializer
func NewUrlHandler(logger *zap.Logger, service *service.UrlService) (*UrlHandler){
	return &UrlHandler{
		logger: logger,
		service: service,
	}
}

func (h *UrlHandler) ShortenURL(c *gin.Context){
	
	userIdRaw, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("user not authenticated")
		
		c.JSON(http.StatusUnauthorized, gin.H{
			"error" : "user not authenticated",
		})
		
		return
	}
	
	userId, ok := userIdRaw.(string)
	if !ok {
		h.logger.Error("invalid user_id type")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid user context",
		})
		return
	}
	
	var req ShortenURLRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid shorten request",
			zap.Error(err),
		)
		
		c.JSON(http.StatusBadRequest, gin.H{
			"error" : "invalid request",
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
		"short_url": base + "/" + code,
	})

} 

func (h *UrlHandler) RedirectURL(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "short code is required"})
		return
	}

	originalURL, err := h.service.GetOriginalURL(code)
	if err != nil {
		h.logger.Error("URL not found or expired", zap.String("code", code), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "url not found or expired"})
		return
	}

	c.Redirect(http.StatusFound, originalURL)
}