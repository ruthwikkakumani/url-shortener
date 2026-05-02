package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ruthwikkakumani/url-shortener/services/redirect-service/internal/service"
	"go.uber.org/zap"
)

type UrlHandler struct {
	logger *zap.Logger
	service *service.UrlService
}

// URL initializer
func NewUrlHandler(logger *zap.Logger, service *service.UrlService) (*UrlHandler){
	return &UrlHandler{
		logger: logger,
		service: service,
	}
}


func (h *UrlHandler) RedirectURL(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "short code is required"})
		return
	}

	originalURL, err := h.service.GetOriginalURL(c.Request.Context(), code)
	if err != nil {
		h.logger.Error("URL not found or expired", zap.String("code", code), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "url not found or expired"})
		return
	}

	c.Redirect(http.StatusFound, originalURL)
}