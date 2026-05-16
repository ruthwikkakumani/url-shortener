package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ruthwikkakumani/redirection-engine/services/url-service/internal/config"
	"github.com/ruthwikkakumani/redirection-engine/services/url-service/internal/service"
	"go.uber.org/zap"
)

type UrlHandler struct {
	logger  *zap.Logger
	service *service.UrlService
}

type ShortenURLRequest struct {
	OriginalURL   string  `json:"original_url" binding:"required"`
	ExpiryMinutes *int    `json:"expiry_minutes" binding:"omitempty,gte=1,lte=43200"`
	CustomCode    *string `json:"custom_code" binding:"omitempty,min=3,max=20"`
}

type UpdateURLRequest struct {
	OriginalURL   *string `json:"original_url" binding:"omitempty"`
	ExpiryMinutes *int    `json:"expiry_minutes" binding:"omitempty,gte=1,lte=43200"`
	CustomCode    *string `json:"custom_code" binding:"omitempty,min=3,max=20"`
}

func NewUrlHandler(logger *zap.Logger, service *service.UrlService) *UrlHandler {
	return &UrlHandler{
		logger:  logger,
		service: service,
	}
}

// ShortenURL godoc
// @Summary Create a short URL
// @Description Shorten a long URL with optional custom code and expiry
// @Tags URL
// @Accept json
// @Produce json
// @Param request body ShortenURLRequest true "URL details"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router / [post]
func (h *UrlHandler) ShortenURL(c *gin.Context) {

	userId := c.GetString("userID")

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

	expiry := 0
	if req.ExpiryMinutes != nil {
		expiry = *req.ExpiryMinutes
	}

	code, err := h.service.CreateShortURL(c.Request.Context(), userId, req.OriginalURL, expiry, req.CustomCode)
	if err != nil {
		h.logger.Error("failed to create short url", zap.Error(err))

		if err.Error() == "short code already in use" || err.Error() == "invalid short code format" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	base := config.GetEnv("BASE_URL", "http://localhost:8082")

	c.JSON(http.StatusCreated, gin.H{
		"short_url": base + "/r/" + code,
	})

}

// ListURLS godoc
// @Summary List all URLs
// @Description Get all URLs created by the authenticated user
// @Tags URL
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router / [get]
func (h *UrlHandler) ListURLS(c *gin.Context) {

	userId := c.GetString("userID")

	urls, err := h.service.ListURLS(c.Request.Context(), userId)
	if err != nil {
		h.logger.Error("unable to process request",
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error" : "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data" : urls,
	})
}

// UpdateURL godoc
// @Summary Update a short URL
// @Description Update the original URL or expiry of an existing short code
// @Tags URL
// @Accept json
// @Produce json
// @Param shortCode path string true "Short Code"
// @Param request body UpdateURLRequest true "Update details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /{shortCode} [put]
func (h *UrlHandler) UpdateURL(c *gin.Context) {

	userId := c.GetString("userID")

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

	shortCode := c.Param("shortCode")

	updatedCode, err := h.service.UpdateURL(c.Request.Context(), userId, req.OriginalURL, shortCode, req.CustomCode, req.ExpiryMinutes)
	if err != nil {
		h.logger.Error("failed to update url", zap.Error(err))

		if err.Error() == "url not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "url not found",
			})
			return
		}

		if err.Error() == "short code already in use" || err.Error() == "invalid short code format" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
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
		"short_url": base + "/r/" + updatedCode,
	})
}


// DeleteURL godoc
// @Summary Delete a short URL
// @Description Delete an existing short URL by its code
// @Tags URL
// @Accept json
// @Produce json
// @Param shortCode path string true "Short Code"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /{shortCode} [delete]
func (h *UrlHandler) DeleteURL(c *gin.Context) {

	userId := c.GetString("userID")

	code := c.Param("shortCode")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error" : "invalid request",
		})
		return
	}

	err := h.service.DeleteURL(c.Request.Context(), userId, code)
	if err != nil {
		h.logger.Error("failed to delete url",
			zap.Error(err),
		)

		if err.Error() == "url not found" {
		    c.JSON(http.StatusNotFound, gin.H{
		        "error": "url not found",
		    })
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error" : "internal server error",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "url deleted successfully",
	})
}